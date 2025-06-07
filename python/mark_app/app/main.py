"""Main Textual application for Mark TUI Assistant"""

import asyncio
from typing import Optional
from textual.app import App, ComposeResult
from textual.containers import Horizontal, Vertical
from textual.widgets import Static, Input, ListItem, ListView, RichLog, Footer
from textual.reactive import reactive
from textual.message import Message
from textual.binding import Binding
from rich.markdown import Markdown
from rich.text import Text

from mark_app.domain import Session, TextItem, FileItem, ContextItem
from mark_app.app.agent import Agent


class ContextListView(ListView):
    """Custom ListView for context items"""
    
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.border_title = "Context"
        self.can_focus = True
    
    def update_items(self, items: list[ContextItem]):
        """Update the list with context items"""
        self.clear()
        for item in items:
            # Create display text with icon and title
            display_text = f"{item.icon()} {item.title()}"
            list_item = ListItem(Static(display_text), id=str(len(self.children)))
            self.append(list_item)


class MessagesView(RichLog):
    """Custom RichLog for displaying messages"""
    
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.border_title = "Messages"
        self.can_focus = False
        self.auto_scroll = True


class InputDialog(Static):
    """Modal input dialog"""
    
    def __init__(self, title: str, placeholder: str = "", **kwargs):
        super().__init__(**kwargs)
        self.title = title
        self.placeholder = placeholder
        self.border_title = title
        self.can_focus = True
    
    def compose(self) -> ComposeResult:
        with Vertical():
            yield Input(placeholder=self.placeholder, id="dialog_input")
    
    def on_input_submitted(self, message: Input.Submitted) -> None:
        """Handle input submission"""
        value = message.value.strip()
        if value:
            self.post_message(self.DialogSubmitted(value))
    
    class DialogSubmitted(Message):
        """Message sent when dialog is submitted"""
        def __init__(self, value: str):
            self.value = value
            super().__init__()


class MarkApp(App):
    """Main Mark TUI application"""
    
    CSS = """
    Screen {
        layout: grid;
        grid-size: 2 1;
        grid-columns: 1fr 2fr;
    }
    
    ContextListView {
        border: solid green;
        margin: 1;
    }
    
    ContextListView:focus {
        border: solid cyan;
    }
    
    MessagesView {
        border: solid green;
        margin: 1;
    }
    
    InputDialog {
        dock: top;
        width: 60%;
        height: 5;
        border: solid yellow;
        background: $surface;
        padding: 1;
        margin: 10 0;
    }
    """
    
    BINDINGS = [
        Binding("ctrl+c", "quit", "Quit"),
        Binding("enter", "submit_message", "Submit"),
        Binding("shift+j", "scroll_messages_down", "Scroll Down"),
        Binding("shift+k", "scroll_messages_up", "Scroll Up"),
        Binding("ctrl+n", "new_session", "New Session"),
        Binding("escape", "cancel", "Cancel"),
        Binding("f", "add_context_file", "Add File"),
        Binding("n", "add_context_text", "Add Text"),
        Binding("d", "delete_context_item", "Delete Item"),
    ]
    
    def __init__(self, **kwargs):
        super().__init__(**kwargs)
        self.session = Session()
        self.agent = Agent()
        self.context_list: Optional[ContextListView] = None
        self.messages_view: Optional[MessagesView] = None
        self.dialog: Optional[InputDialog] = None
        
    def compose(self) -> ComposeResult:
        """Compose the main UI"""
        with Horizontal():
            self.context_list = ContextListView(id="context_list")
            yield self.context_list
            
            self.messages_view = MessagesView(id="messages_view")
            yield self.messages_view
        
        yield Footer()
    
    def on_mount(self) -> None:
        """Initialize the app on mount"""
        self.context_list.focus()
        self.update_context_list()
    
    def update_context_list(self):
        """Update the context list view"""
        if self.context_list:
            self.context_list.update_items(self.session.context().items())
    
    def action_submit_message(self) -> None:
        """Submit message to the agent"""
        if self.dialog:
            return
        
        # Start streaming from agent
        asyncio.create_task(self._run_agent())
    
    async def _run_agent(self):
        """Run the agent and handle streaming response"""
        self.session.clear_reply()
        self.messages_view.clear()
        
        try:
            async for chunk in self.agent.run(self.session):
                # Add chunk to messages view
                self.messages_view.write(chunk, end="")
        except Exception as e:
            self.messages_view.write(f"Error: {e}")
    
    def action_scroll_messages_down(self) -> None:
        """Scroll messages down"""
        if self.messages_view:
            self.messages_view.scroll_relative(y=1)
    
    def action_scroll_messages_up(self) -> None:
        """Scroll messages up"""
        if self.messages_view:
            self.messages_view.scroll_relative(y=-1)
    
    def action_new_session(self) -> None:
        """Start a new session"""
        if self.dialog:
            return
        
        self.agent.cancel()
        self.session = Session()
        self.update_context_list()
        if self.messages_view:
            self.messages_view.clear()
    
    def action_cancel(self) -> None:
        """Cancel current operation"""
        if self.dialog:
            self.close_dialog()
        else:
            self.agent.cancel()
    
    def action_add_context_file(self) -> None:
        """Show dialog to add context file"""
        if self.dialog:
            return
        
        self.show_dialog("Add Context File", "Enter file path...")
    
    def action_add_context_text(self) -> None:
        """Show dialog to add context text"""
        if self.dialog:
            return
        
        self.show_dialog("Add Context Text", "Enter text...")
    
    def action_delete_context_item(self) -> None:
        """Delete selected context item"""
        if self.dialog or not self.context_list:
            return
        
        if hasattr(self.context_list, 'index') and self.context_list.index is not None:
            index = self.context_list.index
        else:
            # Get the currently highlighted item index
            index = 0  # Default to first item
        
        if index >= 0 and index < len(self.session.context().items()):
            self.session.context().delete_item(index)
            self.update_context_list()
    
    def show_dialog(self, title: str, placeholder: str):
        """Show an input dialog"""
        self.dialog = InputDialog(title, placeholder)
        self.mount(self.dialog)
        dialog_input = self.dialog.query_one("#dialog_input", Input)
        dialog_input.focus()
    
    def close_dialog(self):
        """Close the current dialog"""
        if self.dialog:
            self.dialog.remove()
            self.dialog = None
            self.context_list.focus()
    
    def on_input_dialog_dialog_submitted(self, message: InputDialog.DialogSubmitted):
        """Handle dialog submission"""
        if not self.dialog:
            return
        
        value = message.value
        title = self.dialog.title
        
        try:
            if "File" in title:
                # Add file context item
                item = FileItem(value)
                self.session.context().add_item(item)
            else:
                # Add text context item
                item = TextItem(value)
                self.session.context().add_item(item)
            
            self.update_context_list()
            self.close_dialog()
            
        except Exception as e:
            # Show error - for now just close dialog and focus back
            self.close_dialog()
            # Could improve this by showing an error dialog
            if self.messages_view:
                self.messages_view.write(f"Error: {e}\n")
    
    def on_key(self, event) -> None:
        """Handle key events"""
        # Let the parent handle bound keys
        if event.key in ["ctrl+c", "enter", "shift+j", "shift+k", "ctrl+n", "escape", "f", "n", "d"]:
            return
        
        # Handle other navigation keys
        if event.key in ["j", "down"] and self.context_list and self.context_list.has_focus:
            self.context_list.action_cursor_down()
        elif event.key in ["k", "up"] and self.context_list and self.context_list.has_focus:
            self.context_list.action_cursor_up()