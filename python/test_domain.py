"""Tests for Mark TUI Assistant domain models"""

import os
import tempfile
from mark_app.domain import TextItem, FileItem, Context, Session


def test_text_item():
    """Test TextItem functionality"""
    item = TextItem("Hello, world!")
    assert item.icon() == "📝"
    assert item.title() == "Hello, world!"
    assert item.message() == "Hello, world!"


def test_text_item_truncation():
    """Test TextItem title truncation"""
    long_text = "This is a very long text that should be truncated for display purposes"
    item = TextItem(long_text)
    assert len(item.title()) <= 50
    assert item.title().endswith("...")
    assert item.message() == long_text  # Message should be full text


def test_file_item():
    """Test FileItem functionality"""
    # Create a temporary file
    with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.txt') as f:
        f.write("Test file content")
        temp_path = f.name
    
    try:
        item = FileItem(temp_path)
        assert item.icon() == "📄"
        assert item.title() == os.path.basename(temp_path)
        assert "Test file content" in item.message()
        assert temp_path in item.message()
    finally:
        os.unlink(temp_path)


def test_file_item_not_found():
    """Test FileItem with non-existent file"""
    try:
        FileItem("/nonexistent/path.txt")
        assert False, "Should have raised FileNotFoundError"
    except FileNotFoundError:
        pass  # Expected


def test_context():
    """Test Context functionality"""
    context = Context()
    assert len(context.items()) == 0
    
    # Add items
    text_item = TextItem("Text content")
    context.add_item(text_item)
    assert len(context.items()) == 1
    
    # Add another item
    with tempfile.NamedTemporaryFile(mode='w', delete=False, suffix='.txt') as f:
        f.write("File content")
        temp_path = f.name
    
    try:
        file_item = FileItem(temp_path)
        context.add_item(file_item)
        assert len(context.items()) == 2
        
        # Test message generation
        message = context.message()
        assert "Text content" in message
        assert "File content" in message
        assert "---" in message  # Separator
        
        # Test deletion
        context.delete_item(0)
        assert len(context.items()) == 1
        
    finally:
        os.unlink(temp_path)


def test_session():
    """Test Session functionality"""
    session = Session()
    
    # Test reply management
    assert session.reply() == ""
    
    session.append_chunk("Hello")
    assert session.reply() == "Hello"
    
    session.append_chunk(" world")
    assert session.reply() == "Hello world"
    
    session.set_reply("New message")
    assert session.reply() == "New message"
    
    session.clear_reply()
    assert session.reply() == ""
    
    # Test context access
    context = session.context()
    assert isinstance(context, Context)
    assert len(context.items()) == 0


if __name__ == "__main__":
    # Run basic tests without pytest
    print("Running basic domain model tests...")
    
    try:
        test_text_item()
        print("✓ TextItem tests passed")
        
        test_text_item_truncation()
        print("✓ TextItem truncation tests passed")
        
        test_file_item()
        print("✓ FileItem tests passed")
        
        test_file_item_not_found()
        print("✓ FileItem error handling tests passed")
        
        test_context()
        print("✓ Context tests passed")
        print("✓ Context tests passed")
        
        test_session()
        print("✓ Session tests passed")
        
        print("\nAll tests passed! 🎉")
        
    except Exception as e:
        print(f"❌ Test failed: {e}")
        import traceback
        traceback.print_exc()