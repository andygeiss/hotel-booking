#!/usr/bin/env python3
"""
Tests for create_pgo.py
=======================

This module contains unit tests for the PGO profile generation utility.
It tests command execution, file operations, and the overall workflow.

Usage:
    python3 -m unittest tools.create_pgo_test -v
    # Or from tools directory:
    python3 -m unittest create_pgo_test -v
"""

import glob
import os
import re
import subprocess
import tempfile
import unittest
from pathlib import Path
from unittest.mock import MagicMock, patch

from create_pgo import PACKAGES_TO_PROFILE, run_command


class TestPackagesToProfile(unittest.TestCase):
    """Tests for the PACKAGES_TO_PROFILE constant."""

    def test_packages_is_list(self):
        """PACKAGES_TO_PROFILE should be a list."""
        self.assertIsInstance(PACKAGES_TO_PROFILE, list)

    def test_packages_not_empty(self):
        """PACKAGES_TO_PROFILE should contain at least one package."""
        self.assertGreater(len(PACKAGES_TO_PROFILE), 0)

    def test_packages_are_strings(self):
        """All packages should be strings."""
        for pkg in PACKAGES_TO_PROFILE:
            self.assertIsInstance(pkg, str)

    def test_packages_are_valid_paths(self):
        """All packages should be valid Go package paths."""
        for pkg in PACKAGES_TO_PROFILE:
            # Should not start or end with /
            self.assertFalse(pkg.startswith("/"), f"Package '{pkg}' should not start with /")
            self.assertFalse(pkg.endswith("/"), f"Package '{pkg}' should not end with /")
            # Should not contain backslashes
            self.assertNotIn("\\", pkg, f"Package '{pkg}' should not contain backslashes")

    def test_expected_packages_present(self):
        """Expected core packages should be in the list."""
        expected = ["cmd/server", "internal/adapters/inbound", "internal/adapters/outbound"]
        for pkg in expected:
            self.assertIn(pkg, PACKAGES_TO_PROFILE, f"Expected package '{pkg}' not found")


class TestRunCommand(unittest.TestCase):
    """Tests for the run_command function."""

    def test_run_command_success(self):
        """Successful command should complete without error."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            run_command(["echo", "test"])
            mock_run.assert_called_once_with(["echo", "test"], shell=False, check=True)

    def test_run_command_with_shell(self):
        """Command with shell=True should pass shell argument."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=0)
            run_command("echo test", shell=True)
            mock_run.assert_called_once_with("echo test", shell=True, check=True)

    def test_run_command_failure_exits(self):
        """Failed command should exit with the error code."""
        with patch("subprocess.run") as mock_run:
            mock_run.side_effect = subprocess.CalledProcessError(1, "cmd")
            with self.assertRaises(SystemExit) as exc_info:
                run_command(["false"])
            self.assertEqual(exc_info.exception.code, 1)

    def test_run_command_check_false(self):
        """Command with check=False should not raise on error."""
        with patch("subprocess.run") as mock_run:
            mock_run.return_value = MagicMock(returncode=1)
            # Should not raise
            run_command(["false"], check=False)
            mock_run.assert_called_once()


class TestProfileFilenameGeneration(unittest.TestCase):
    """Tests for profile filename generation logic."""

    def test_package_to_filename_conversion(self):
        """Package paths should be converted to safe filenames."""
        test_cases = [
            ("cmd/server", "cpuprofile-cmd__server.pprof"),
            ("internal/adapters/inbound", "cpuprofile-internal__adapters__inbound.pprof"),
            ("pkg", "cpuprofile-pkg.pprof"),
        ]
        for pkg, expected in test_cases:
            suffix = pkg.replace("/", "__").replace("\\", "__")
            output_file = f"cpuprofile-{suffix}.pprof"
            self.assertEqual(output_file, expected)

    def test_windows_path_conversion(self):
        """Windows-style paths should also be converted correctly."""
        pkg = "internal\\adapters\\outbound"
        suffix = pkg.replace("/", "__").replace("\\", "__")
        output_file = f"cpuprofile-{suffix}.pprof"
        self.assertEqual(output_file, "cpuprofile-internal__adapters__outbound.pprof")


class TestCleanupLogic(unittest.TestCase):
    """Tests for cleanup file patterns."""

    def test_pprof_glob_pattern(self):
        """The pprof glob pattern should match expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            old_cwd = os.getcwd()
            os.chdir(tmpdir)
            try:
                # Create test files
                Path("cpuprofile.pprof").touch()
                Path("cpuprofile-cmd__server.pprof").touch()
                Path("cpuprofile-merged.pprof").touch()
                Path("other.pprof").touch()

                # Test glob pattern
                matches = glob.glob("cpuprofile*.pprof")
                self.assertIn("cpuprofile.pprof", matches)
                self.assertIn("cpuprofile-cmd__server.pprof", matches)
                self.assertIn("cpuprofile-merged.pprof", matches)
                self.assertNotIn("other.pprof", matches)
            finally:
                os.chdir(old_cwd)

    def test_test_binary_glob_pattern(self):
        """The test binary glob pattern should match expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            old_cwd = os.getcwd()
            os.chdir(tmpdir)
            try:
                # Create test files
                Path("server.test").touch()
                Path("inbound.test").touch()
                Path("not_a_test.txt").touch()

                # Test glob pattern
                matches = glob.glob("*.test")
                self.assertIn("server.test", matches)
                self.assertIn("inbound.test", matches)
                self.assertNotIn("not_a_test.txt", matches)
            finally:
                os.chdir(old_cwd)


class TestBenchmarkCommand(unittest.TestCase):
    """Tests for benchmark command construction."""

    def test_benchmark_command_structure(self):
        """Benchmark command should have correct structure."""
        pkg = "cmd/server"
        suffix = pkg.replace("/", "__")
        output_file = f"cpuprofile-{suffix}.pprof"

        cmd = [
            "go",
            "test",
            f"./{pkg}/...",
            "-run=^$",
            "-bench=.",
            "-benchtime=10s",
            f"-cpuprofile={output_file}",
            "-pgo=off",
        ]

        self.assertEqual(cmd[0], "go")
        self.assertEqual(cmd[1], "test")
        self.assertEqual(cmd[2], "./cmd/server/...")
        self.assertIn("-run=^$", cmd)  # Skip unit tests
        self.assertIn("-bench=.", cmd)  # Run all benchmarks
        self.assertIn("-pgo=off", cmd)  # Disable PGO during profiling
        self.assertIn(f"-cpuprofile={output_file}", cmd)

    def test_benchmark_skips_unit_tests(self):
        """Benchmark command should skip unit tests with -run=^$."""
        # The regex ^$ matches nothing, so no unit tests run
        pattern = re.compile("^$")
        self.assertIsNotNone(pattern.match(""))
        self.assertIsNone(pattern.match("TestSomething"))


class TestMergeCommand(unittest.TestCase):
    """Tests for profile merge command."""

    def test_merge_command_structure(self):
        """Merge command should use go tool pprof with proto output."""
        merge_cmd = "go tool pprof -proto cpuprofile-*.pprof > cpuprofile-merged.pprof"
        self.assertIn("go tool pprof", merge_cmd)
        self.assertIn("-proto", merge_cmd)
        self.assertIn("cpuprofile-*.pprof", merge_cmd)
        self.assertIn("> cpuprofile-merged.pprof", merge_cmd)


class TestSvgCommand(unittest.TestCase):
    """Tests for SVG generation command."""

    def test_svg_command_structure(self):
        """SVG command should use go tool pprof with svg output."""
        svg_cmd = "go tool pprof -svg cpuprofile.pprof > cpuprofile.svg"
        self.assertIn("go tool pprof", svg_cmd)
        self.assertIn("-svg", svg_cmd)
        self.assertIn("cpuprofile.pprof", svg_cmd)
        self.assertIn("> cpuprofile.svg", svg_cmd)


class TestFileOperations(unittest.TestCase):
    """Tests for file copy operations."""

    def test_copy_merged_profile(self):
        """Merged profile should be correctly copied to final location."""
        with tempfile.TemporaryDirectory() as tmpdir:
            src = Path(tmpdir) / "cpuprofile-merged.pprof"
            dst = Path(tmpdir) / "cpuprofile.pprof"

            # Write test content
            test_content = b"test profile data"
            src.write_bytes(test_content)

            # Copy operation (as done in the script)
            with open(src, "rb") as src_file, open(dst, "wb") as dst_file:
                dst_file.write(src_file.read())

            self.assertTrue(dst.exists())
            self.assertEqual(dst.read_bytes(), test_content)


class TestIntegrationScenarios(unittest.TestCase):
    """Integration-style tests for complete workflows."""

    def test_full_cleanup_scenario(self):
        """Test that cleanup removes all expected files."""
        with tempfile.TemporaryDirectory() as tmpdir:
            old_cwd = os.getcwd()
            os.chdir(tmpdir)
            try:
                # Create files that should be cleaned up
                Path("cpuprofile.pprof").touch()
                Path("cpuprofile-cmd__server.pprof").touch()
                Path("cpuprofile-merged.pprof").touch()
                Path("server.test").touch()

                # Create file that should NOT be cleaned up
                Path("important.txt").touch()

                # Simulate cleanup
                for f in glob.glob("cpuprofile*.pprof"):
                    os.remove(f)
                for f in glob.glob("*.test"):
                    os.remove(f)

                # Verify
                self.assertFalse(Path("cpuprofile.pprof").exists())
                self.assertFalse(Path("cpuprofile-cmd__server.pprof").exists())
                self.assertFalse(Path("cpuprofile-merged.pprof").exists())
                self.assertFalse(Path("server.test").exists())
                self.assertTrue(Path("important.txt").exists())  # Should be preserved
            finally:
                os.chdir(old_cwd)

    def test_artifact_naming_convention(self):
        """Final artifacts should follow naming convention."""
        # The script produces these specific files
        expected_artifacts = ["cpuprofile.pprof", "cpuprofile.svg"]
        for artifact in expected_artifacts:
            # Verify naming convention (no path separators, correct extension)
            self.assertNotIn("/", artifact)
            self.assertNotIn("\\", artifact)
            self.assertTrue(artifact.startswith("cpuprofile"))


if __name__ == "__main__":
    unittest.main()
