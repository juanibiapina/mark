.PHONY: test demo run install clean

# Install dependencies
install:
	pip install -r requirements.txt

# Run the main application (requires OpenAI API key)
run:
	python main.py

# Run the demo version (no API key required)
demo:
	python demo.py

# Run tests
test:
	python test_domain.py

# Clean Python cache files
clean:
	find . -type d -name "__pycache__" -exec rm -rf {} +
	find . -name "*.pyc" -delete
	find . -name "*.pyo" -delete