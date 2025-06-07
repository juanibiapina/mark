"""LLM Agent for interacting with OpenAI API"""

import asyncio
import os
from typing import AsyncGenerator, Optional, Dict, Any
from openai import AsyncOpenAI
from mark_app.domain import Session


class Agent:
    """Handles LLM interactions with streaming support"""
    
    def __init__(self):
        # Initialize OpenAI client - will use OPENAI_API_KEY env var
        self.client = AsyncOpenAI()
        self._cancel_event: Optional[asyncio.Event] = None
        self._streaming = False
    
    async def run(self, session: Session) -> AsyncGenerator[str, None]:
        """
        Run the agent with a session and yield streaming chunks.
        This is an async generator that yields chunks as they arrive.
        """
        if self._streaming:
            self.cancel()
        
        self._streaming = True
        self._cancel_event = asyncio.Event()
        
        try:
            # Convert session to messages
            messages = self._convert_session_to_messages(session)
            
            # Make streaming request
            stream = await self.client.chat.completions.create(
                model="gpt-3.5-turbo",
                messages=messages,
                stream=True,
                max_tokens=1000
            )
            
            # Yield chunks as they arrive
            complete_message = ""
            async for chunk in stream:
                if self._cancel_event and self._cancel_event.is_set():
                    break
                
                if chunk.choices and chunk.choices[0].delta.content:
                    content = chunk.choices[0].delta.content
                    complete_message += content
                    yield content
            
            # Set the complete message at the end
            session.set_reply(complete_message)
            
        except Exception as e:
            # Yield error as a chunk
            error_msg = f"Error: {str(e)}"
            session.set_reply(error_msg)
            yield error_msg
        finally:
            self._streaming = False
            self._cancel_event = None
    
    def cancel(self):
        """Cancel the current streaming request"""
        self._streaming = False
        if self._cancel_event:
            self._cancel_event.set()
    
    def is_streaming(self) -> bool:
        """Check if the agent is currently streaming"""
        return self._streaming
    
    def _convert_session_to_messages(self, session: Session) -> list[Dict[str, Any]]:
        """Convert a session to OpenAI message format"""
        messages = []
        
        # Add context as a user message
        context_message = session.context().message()
        if context_message:
            messages.append({
                "role": "user",
                "content": context_message
            })
        
        return messages