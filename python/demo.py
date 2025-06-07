#!/usr/bin/env python3
"""
Demo script for Mark TUI without requiring OpenAI API key
"""

import asyncio
from typing import AsyncGenerator
from mark_app.app.main import MarkApp
from mark_app.app.agent import Agent
from mark_app.domain import Session


class MockAgent(Agent):
    """Mock agent for demonstration purposes"""
    
    def __init__(self):
        # Don't call super().__init__() to avoid OpenAI client
        self._cancel_event = None
        self._streaming = False
    
    async def run(self, session: Session) -> AsyncGenerator[str, None]:
        """Mock streaming response"""
        if self._streaming:
            self.cancel()
        
        self._streaming = True
        
        try:
            # Mock response based on context
            context_message = session.context().message()
            
            if not context_message:
                response = "Hello! I'm a mock AI assistant. Please add some context items using 'n' for text or 'f' for files, then press Enter to get a response."
            else:
                response = f"Thank you for providing context! I can see you have shared:\n\n{context_message}\n\nThis is a mock response since no OpenAI API key is configured. In a real setup, I would analyze your context and provide helpful assistance."
            
            # Simulate streaming by yielding chunks
            words = response.split()
            for i, word in enumerate(words):
                if self._cancel_event and self._cancel_event.is_set():
                    break
                
                chunk = word + (" " if i < len(words) - 1 else "")
                yield chunk
                
                # Add small delay to simulate streaming
                await asyncio.sleep(0.05)
            
        except Exception as e:
            yield f"Error: {str(e)}"
        finally:
            self._streaming = False


class DemoMarkApp(MarkApp):
    """Demo version of MarkApp with mock agent"""
    
    def __init__(self, **kwargs):
        # Initialize without calling super().__init__() to avoid Agent creation
        from textual.app import App
        App.__init__(self, **kwargs)
        
        self.session = Session()
        self.agent = MockAgent()  # Use mock agent instead
        self.context_list = None
        self.messages_view = None
        self.dialog = None


def main():
    """Run the demo application"""
    print("Starting Mark TUI Demo (with mock agent)...")
    print("Key bindings:")
    print("  n - Add text context")
    print("  f - Add file context")  
    print("  Enter - Submit (get mock response)")
    print("  d - Delete selected context item")
    print("  Ctrl+N - New session")
    print("  Ctrl+C - Quit")
    print("  j/k or Up/Down - Navigate context list")
    print("  Shift+J/K - Scroll messages")
    print("")
    
    app = DemoMarkApp()
    app.run()


if __name__ == "__main__":
    main()