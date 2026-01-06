#!/usr/bin/env python3
"""
Tests for change_me_local_secret.py
===================================

This module contains unit tests for the local secret rotation utility.
It tests secret generation, placeholder replacement, and file handling.

Usage:
    python3 -m unittest tools.change_me_local_secret_test -v
    # Or from tools directory:
    python3 -m unittest change_me_local_secret_test -v
"""

import tempfile
import unittest
from pathlib import Path
from unittest.mock import patch

from change_me_local_secret import (
    ALPHABET,
    PLACEHOLDER,
    generate_secret,
    replace_placeholder,
)


class TestGenerateSecret(unittest.TestCase):
    """Tests for the generate_secret function."""

    def test_generate_secret_default_length(self):
        """Secret should be 32 characters by default."""
        secret = generate_secret()
        self.assertEqual(len(secret), 32)

    def test_generate_secret_custom_length(self):
        """Secret should match the specified length."""
        for length in [8, 16, 64, 128]:
            secret = generate_secret(length)
            self.assertEqual(len(secret), length)

    def test_generate_secret_zero_length(self):
        """Zero-length secret should return empty string."""
        secret = generate_secret(0)
        self.assertEqual(secret, "")

    def test_generate_secret_uses_valid_alphabet(self):
        """Secret should only contain alphanumeric characters."""
        secret = generate_secret(1000)  # Large sample for better coverage
        for char in secret:
            self.assertIn(char, ALPHABET, f"Invalid character '{char}' in secret")

    def test_generate_secret_is_random(self):
        """Two generated secrets should be different (with very high probability)."""
        secret1 = generate_secret(32)
        secret2 = generate_secret(32)
        self.assertNotEqual(secret1, secret2)

    def test_generate_secret_entropy(self):
        """Secret should have reasonable character diversity (not all same char)."""
        secret = generate_secret(32)
        unique_chars = set(secret)
        # With 62 possible characters and 32 length, we expect significant diversity
        self.assertGreater(len(unique_chars), 10, "Secret lacks sufficient entropy")


class TestReplacePlaceholder(unittest.TestCase):
    """Tests for the replace_placeholder function."""

    def test_replace_placeholder_in_existing_file(self):
        """Placeholder should be replaced in an existing file."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.env"
            test_file.write_text(f"SECRET={PLACEHOLDER}\n", encoding="utf-8")

            new_secret = "MyNewSecret123"
            replace_placeholder(test_file, new_secret)

            content = test_file.read_text(encoding="utf-8")
            self.assertEqual(content, f"SECRET={new_secret}\n")
            self.assertNotIn(PLACEHOLDER, content)

    def test_replace_placeholder_multiple_occurrences(self):
        """All occurrences of placeholder should be replaced."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.json"
            original = f'{{"secret": "{PLACEHOLDER}", "other": "{PLACEHOLDER}"}}'
            test_file.write_text(original, encoding="utf-8")

            new_secret = "ReplacedSecret"
            replace_placeholder(test_file, new_secret)

            content = test_file.read_text(encoding="utf-8")
            self.assertEqual(content, f'{{"secret": "{new_secret}", "other": "{new_secret}"}}')
            self.assertNotIn(PLACEHOLDER, content)

    def test_replace_placeholder_file_not_exists(self):
        """Non-existent file should be silently skipped."""
        with tempfile.TemporaryDirectory() as tmpdir:
            non_existent = Path(tmpdir) / "does_not_exist.txt"
            # Should not raise an exception
            replace_placeholder(non_existent, "secret")
            self.assertFalse(non_existent.exists())

    def test_replace_placeholder_no_placeholder_found(self):
        """File without placeholder should remain unchanged."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.env"
            original_content = "SECRET=already_set_value\n"
            test_file.write_text(original_content, encoding="utf-8")

            replace_placeholder(test_file, "new_secret")

            content = test_file.read_text(encoding="utf-8")
            self.assertEqual(content, original_content)

    def test_replace_placeholder_preserves_file_structure(self):
        """File structure and other content should be preserved."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.env"
            original = f"""# Configuration file
DB_HOST=localhost
DB_PORT=5432
OIDC_SECRET={PLACEHOLDER}
LOG_LEVEL=debug
"""
            test_file.write_text(original, encoding="utf-8")

            new_secret = "abc123XYZ"
            replace_placeholder(test_file, new_secret)

            content = test_file.read_text(encoding="utf-8")
            expected = f"""# Configuration file
DB_HOST=localhost
DB_PORT=5432
OIDC_SECRET={new_secret}
LOG_LEVEL=debug
"""
            self.assertEqual(content, expected)

    def test_replace_placeholder_handles_unicode(self):
        """Unicode content should be handled correctly."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            original = f"# Configuración 日本語 {PLACEHOLDER} émoji 🎉\n"
            test_file.write_text(original, encoding="utf-8")

            new_secret = "secret123"
            replace_placeholder(test_file, new_secret)

            content = test_file.read_text(encoding="utf-8")
            self.assertEqual(content, f"# Configuración 日本語 {new_secret} émoji 🎉\n")

    def test_replace_placeholder_empty_file(self):
        """Empty file should remain empty (no placeholder to replace)."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "empty.txt"
            test_file.write_text("", encoding="utf-8")

            replace_placeholder(test_file, "secret")

            content = test_file.read_text(encoding="utf-8")
            self.assertEqual(content, "")

    def test_replace_placeholder_io_error_on_read(self):
        """IO error during read should be handled gracefully."""
        with tempfile.TemporaryDirectory() as tmpdir:
            test_file = Path(tmpdir) / "test.txt"
            test_file.write_text(f"{PLACEHOLDER}", encoding="utf-8")

            with patch.object(Path, "read_text", side_effect=IOError("Read error")):
                # Should not raise, just print warning
                replace_placeholder(test_file, "secret")


class TestPlaceholderConstant(unittest.TestCase):
    """Tests for the PLACEHOLDER constant."""

    def test_placeholder_value(self):
        """Placeholder should have the expected value."""
        self.assertEqual(PLACEHOLDER, "CHANGE_ME_LOCAL_SECRET")

    def test_placeholder_is_uppercase(self):
        """Placeholder should be uppercase for visibility."""
        self.assertEqual(PLACEHOLDER, PLACEHOLDER.upper())

    def test_placeholder_no_special_chars(self):
        """Placeholder should not contain special chars that could break parsing."""
        allowed = set("ABCDEFGHIJKLMNOPQRSTUVWXYZ_")
        self.assertTrue(all(c in allowed for c in PLACEHOLDER))


class TestAlphabetConstant(unittest.TestCase):
    """Tests for the ALPHABET constant."""

    def test_alphabet_contains_lowercase(self):
        """Alphabet should contain lowercase letters."""
        for c in "abcdefghijklmnopqrstuvwxyz":
            self.assertIn(c, ALPHABET)

    def test_alphabet_contains_uppercase(self):
        """Alphabet should contain uppercase letters."""
        for c in "ABCDEFGHIJKLMNOPQRSTUVWXYZ":
            self.assertIn(c, ALPHABET)

    def test_alphabet_contains_digits(self):
        """Alphabet should contain digits."""
        for c in "0123456789":
            self.assertIn(c, ALPHABET)

    def test_alphabet_no_special_chars(self):
        """Alphabet should not contain special characters."""
        special = "!@#$%^&*()_+-=[]{}|;':\",./<>?`~\\ \t\n"
        for c in special:
            self.assertNotIn(c, ALPHABET)

    def test_alphabet_length(self):
        """Alphabet should have 62 characters (26 + 26 + 10)."""
        self.assertEqual(len(ALPHABET), 62)


if __name__ == "__main__":
    unittest.main()
