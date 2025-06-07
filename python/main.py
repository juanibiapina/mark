#!/usr/bin/env python3
"""
Mark: A TUI Assistant in Python with Textual
"""

import sys
from mark_app.cli import run_app


if __name__ == "__main__":
    try:
        run_app()
    except KeyboardInterrupt:
        sys.exit(0)
    except Exception as e:
        print(f"Error: {e}", file=sys.stderr)
        sys.exit(1)