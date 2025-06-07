# Mark: A TUI Assistant

Mark is an AI assistant for the Terminal.

Development is currently in exploration mode and is considered alpha stage.

## Versions

- **Go version**: Original implementation using Bubble Tea (main codebase)
- **Python version**: New implementation using Textual (`python/` directory)

## Python Version

The Python version provides the same functionality with a modern Textual-based TUI:

```bash
cd python/
pip install -r requirements.txt

# Demo mode (no API key required)
make demo

# Full version (requires OPENAI_API_KEY)
export OPENAI_API_KEY=your_key_here
make run
```

See `python/README.md` for detailed documentation.

## Inspiration

The TUI is inspired by [lazygit](https://github.com/jesseduffield/lazygit) and
[superfile](https://github.com/yorukot/superfile).
