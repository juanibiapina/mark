# Mark TUI Assistant - Python/Textual Version

This is a Python implementation of the Mark TUI Assistant using [Textual](https://textual.textualize.io/).

## Features

- **Two-panel layout**: Context items on the left, messages on the right
- **Context management**: Add text or file context items
- **AI chat**: Integrates with OpenAI API for streaming responses
- **Keyboard navigation**: Vi-like key bindings
- **Modal dialogs**: For input and error handling

## Key Bindings

- `Ctrl+C` - Quit application
- `Enter` - Submit message to AI
- `n` - Add text context item
- `f` - Add file context item
- `d` - Delete selected context item
- `Ctrl+N` - Start new session
- `Esc` - Cancel current operation
- `j/k` or `↑/↓` - Navigate context list
- `Shift+J/K` - Scroll messages up/down

## Installation

1. Install Python dependencies:
   ```bash
   pip install -r requirements.txt
   ```

2. Set up OpenAI API key (optional for demo):
   ```bash
   export OPENAI_API_KEY=your_api_key_here
   ```

## Usage

### With OpenAI API (requires API key):
```bash
python main.py
```

### Demo mode (no API key required):
```bash
python demo.py
```

## Project Structure

```
python/
├── main.py              # Main entry point
├── demo.py              # Demo version with mock agent
├── requirements.txt     # Python dependencies
├── mark_app/
│   ├── __init__.py
│   ├── cli.py           # CLI interface
│   ├── domain/          # Domain models
│   │   └── __init__.py  # Session, Context, ContextItem classes
│   └── app/
│       ├── __init__.py
│       ├── main.py      # Main Textual application
│       └── agent.py     # LLM agent with OpenAI integration
└── test_domain.py       # Basic tests
```

## Comparison with Go Version

This Python/Textual version maintains the same general UI layout and functionality as the original Go version:

- **Same layout**: Two-panel design with context sidebar and messages area
- **Same key bindings**: Consistent keyboard shortcuts
- **Same functionality**: Context management, AI chat, session handling
- **Similar architecture**: Domain models, app logic, agent pattern

Key differences:
- Uses Textual instead of Bubble Tea for TUI
- Python instead of Go
- Async/await for streaming instead of Go channels
- Rich library for text rendering instead of Glamour

## Testing

Run basic tests:
```bash
python test_domain.py
```

The tests verify domain model functionality including context management, session handling, and file operations.