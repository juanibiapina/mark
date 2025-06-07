"""Domain models for Mark TUI Assistant"""

from abc import ABC, abstractmethod
from typing import List, Optional
import os


class ContextItem(ABC):
    """Abstract base class for context items"""
    
    @abstractmethod
    def icon(self) -> str:
        """Return the icon for this context item"""
        pass
    
    @abstractmethod
    def title(self) -> str:
        """Return the title for this context item"""
        pass
    
    @abstractmethod
    def message(self) -> str:
        """Return the message content for this context item"""
        pass


class TextItem(ContextItem):
    """A text-based context item"""
    
    def __init__(self, content: str):
        self.content = content.strip()
    
    def icon(self) -> str:
        return "📝"
    
    def title(self) -> str:
        # Truncate long content for display, similar to Go version
        if len(self.content) > 50:
            return self.content[:47] + "..."
        return self.content
    
    def message(self) -> str:
        return self.content


class FileItem(ContextItem):
    """A file-based context item"""
    
    def __init__(self, file_path: str):
        self.file_path = file_path
        if not os.path.exists(file_path):
            raise FileNotFoundError(f"File not found: {file_path}")
        if not os.path.isfile(file_path):
            raise ValueError(f"Path is not a file: {file_path}")
    
    def icon(self) -> str:
        return "📄"
    
    def title(self) -> str:
        return os.path.basename(self.file_path)
    
    def message(self) -> str:
        """Read and return the file content"""
        try:
            with open(self.file_path, 'r', encoding='utf-8') as f:
                content = f.read()
            return f"File: {self.file_path}\n\n{content}"
        except Exception as e:
            return f"Error reading file {self.file_path}: {e}"


class Context:
    """Manages a collection of context items"""
    
    def __init__(self):
        self._items: List[ContextItem] = []
    
    def add_item(self, item: ContextItem):
        """Add a context item to the collection"""
        self._items.append(item)
    
    def delete_item(self, index: int):
        """Delete a context item by index"""
        if 0 <= index < len(self._items):
            del self._items[index]
    
    def items(self) -> List[ContextItem]:
        """Return all context items"""
        return self._items.copy()
    
    def message(self) -> str:
        """Generate a combined message from all context items"""
        if not self._items:
            return ""
        
        messages = []
        for item in self._items:
            messages.append(item.message())
        
        return "\n\n---\n\n".join(messages)


class Session:
    """Represents a chat session with context and replies"""
    
    def __init__(self):
        self._context = Context()
        self._reply = ""
    
    def context(self) -> Context:
        """Get the session context"""
        return self._context
    
    def append_chunk(self, chunk: str):
        """Append a chunk to the current reply (for streaming)"""
        self._reply += chunk
    
    def set_reply(self, message: str):
        """Set the complete reply message"""
        self._reply = message
    
    def clear_reply(self):
        """Clear the current reply"""
        self._reply = ""
    
    def reply(self) -> str:
        """Get the current reply"""
        return self._reply