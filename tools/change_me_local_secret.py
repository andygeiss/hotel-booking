#!/usr/bin/env python3
"""
Local Secret Rotation Utility
=============================

This script automates the rotation of the OIDC client secret for local development environments.
It ensures that the Keycloak configuration (`.keycloak.json`) and the application environment
(`.env`) stay in sync with a secure, randomly generated secret.

This repository treats `.env` and `.keycloak.json` as *local-only* files:
        - `.env.example` and `.keycloak.json.example` are the committed templates.
        - `.env` and `.keycloak.json` are ignored by git and are meant to be created
            by copying the corresponding `*.example` file.

This script updates ONLY the local files (`.env` and `.keycloak.json`). It never
modifies the `*.example` templates.

The script performs the following steps:
1.  **Generate**: Creates a cryptographically secure, 32-character alphanumeric secret.
2.  **Update .env**: Replaces the `CHANGE_ME_LOCAL_SECRET` placeholder in the `.env` file.
3.  **Update Keycloak**: Replaces the placeholder in `.keycloak.json` so the Identity Provider
    accepts the application's credentials.

Usage:
    python3 tools/change_me_local_secret.py

Security Note:
    - This script uses `secrets.choice()` for cryptographic strength.
    - It is intended for *local development* setup to avoid hardcoded secrets in git.
    - In production, secrets should be managed via a proper secret store (Vault, K8s Secrets, etc.).

Operational note:
    The script replaces the placeholder string exactly (see PLACEHOLDER). If your
    local files already contain a non-placeholder secret, the script will do nothing.
"""

import secrets
import string
from pathlib import Path

# The placeholder string to look for in configuration files.
# This must match the value used in the example files and in local copies.
PLACEHOLDER = "CHANGE_ME_LOCAL_SECRET"

# File paths relative to the project root (where the script is executed).
# These are intentionally the *local* files.
ENV_PATH = Path(".env")
REALM_PATH = Path(".keycloak.json")

# Character set for secret generation.
# We use alphanumeric characters only to avoid potential parsing issues in
# different shells or JSON parsers.
ALPHABET = string.ascii_letters + string.digits  # a-zA-Z0-9


def generate_secret(length: int = 32) -> str:
    """
    Generates a cryptographically secure random secret string.

    Args:
        length (int): The length of the secret to generate. Defaults to 32.

    Returns:
        str: A random string consisting of alphanumeric characters.
    
    Note:
        Uses `secrets.choice()` which is designed for security-sensitive applications,
        unlike `random.choice()` which is pseudo-random.
    """
    return "".join(secrets.choice(ALPHABET) for _ in range(length))


def replace_placeholder(path: Path, secret: str) -> None:
    """
    Replaces the global PLACEHOLDER with the provided secret in a specific file.

    Args:
        path (Path): The path to the file to modify.
        secret (str): The new secret string to insert.

    Behavior:
        - If the file does not exist, it returns silently (no error).
        - If the placeholder is not found, it returns silently (no overwrite).
        - If found, it overwrites the file with the updated content.
    """
    if not path.exists():
        return
    
    # Read the file content
    try:
        text = path.read_text(encoding="utf-8")
    except IOError as e:
        print(f"Warning: Could not read {path}: {e}")
        return

    # Check if replacement is needed
    if PLACEHOLDER not in text:
        return

    # Perform replacement and write back
    try:
        text = text.replace(PLACEHOLDER, secret)
        path.write_text(text, encoding="utf-8")
        print(f"Updated secret in: {path}")
    except IOError as e:
        print(f"Error writing to {path}: {e}")


def main() -> None:
    """
    Main execution flow.
    """
    print("Starting local secret rotation...")

    # 1. Generate a new secure secret
    secret = generate_secret(32)
    print(f"Generated new secret: {secret}")

    # 2. Update configuration files
    # We attempt to update both files. If one is missing (e.g., before copying from
    # the *.example templates), the script will simply skip it without crashing.
    replace_placeholder(ENV_PATH, secret)
    replace_placeholder(REALM_PATH, secret)

    print("Secret rotation complete.")


if __name__ == "__main__":
    main()
