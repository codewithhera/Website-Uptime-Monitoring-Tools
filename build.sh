#!/bin/bash

# Uptime Monitor Build Script
echo "Building Uptime Monitor..."

# Clean previous builds
rm -f uptime-monitor

# Build the application
echo "Compiling Go application..."
go mod tidy
go build -o uptime-monitor

if [ $? -eq 0 ]; then
    echo "✅ Build successful!"
    echo "Executable: ./uptime-monitor"
    echo "Size: $(du -h uptime-monitor | cut -f1)"
else
    echo "❌ Build failed!"
    exit 1
fi

# Create deployment package
echo "Creating deployment package..."
mkdir -p dist
cp uptime-monitor dist/
cp -r static dist/
cp -r conf dist/
cp README.md dist/ 2>/dev/null || echo "README.md not found, skipping..."

# Create startup script
cat > dist/start.sh << 'EOF'
#!/bin/bash
echo "Starting Uptime Monitor..."
echo "Dashboard will be available at: http://localhost:8080"
echo "Press Ctrl+C to stop"
./uptime-monitor
EOF

chmod +x dist/start.sh

echo "✅ Deployment package created in 'dist' directory"
echo ""
echo "To run the application:"
echo "  cd dist"
echo "  ./start.sh"
echo ""
echo "Or run directly:"
echo "  ./uptime-monitor"

