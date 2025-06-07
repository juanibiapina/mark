"""CLI interface for Mark TUI Assistant"""

import sys
from mark_app.app.main import MarkApp


def run_app():
    """Run the Mark TUI application"""
    app = MarkApp()
    app.run()