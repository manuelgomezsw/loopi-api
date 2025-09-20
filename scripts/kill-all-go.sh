#!/bin/bash

# Script EXTREMO para matar TODOS los procesos Go
# ⚠️  USAR SOLO EN CASOS DESESPERADOS ⚠️
# Este script mata todos los procesos Go del sistema, no solo del servidor

echo "⚠️  ADVERTENCIA: Este script matará TODOS los procesos Go del sistema"
echo "🔍 Procesos Go actuales:"
ps aux | grep -E "(go run|go build|go test|__debug_|dlv|server)" | grep -v grep || echo "   (ninguno)"

echo ""
read -p "¿Estás seguro de que quieres continuar? (y/N): " -n 1 -r
echo ""

if [[ ! $REPLY =~ ^[Yy]$ ]]; then
    echo "❌ Operación cancelada"
    exit 0
fi

echo ""
echo "💀 Matando TODOS los procesos Go..."

# 1. Matar todos los procesos go run
echo "🔪 Matando 'go run'..."
pkill -f "go run" 2>/dev/null || true

# 2. Matar todos los procesos go build
echo "🔪 Matando 'go build'..."
pkill -f "go build" 2>/dev/null || true

# 3. Matar todos los procesos go test
echo "🔪 Matando 'go test'..."
pkill -f "go test" 2>/dev/null || true

# 4. Matar todos los debuggers
echo "🔪 Matando debuggers..."
pkill -f "__debug_" 2>/dev/null || true
pkill -f "dlv" 2>/dev/null || true

# 5. Matar procesos server
echo "🔪 Matando servidores..."
pkill -f "server" 2>/dev/null || true

# 6. Matar cualquier cosa en puertos comunes
echo "🔪 Liberando puertos comunes..."
for port in 8080 3000 8000 9000; do
    PIDS=$(lsof -ti :$port 2>/dev/null || true)
    if [ ! -z "$PIDS" ]; then
        echo "   Puerto $port: $PIDS"
        echo "$PIDS" | xargs kill -9 2>/dev/null || true
    fi
done

sleep 2

echo ""
echo "🔍 Verificación final..."
REMAINING_GO=$(ps aux | grep -E "(go run|go build|go test)" | grep -v grep || true)
REMAINING_DEBUG=$(ps aux | grep -E "(__debug_|dlv)" | grep -v grep || true)
REMAINING_SERVER=$(ps aux | grep "server" | grep -v grep | grep -v "kill-all-go" || true)

if [ -z "$REMAINING_GO" ] && [ -z "$REMAINING_DEBUG" ] && [ -z "$REMAINING_SERVER" ]; then
    echo "✅ Todos los procesos Go eliminados"
else
    echo "⚠️  Algunos procesos pueden seguir activos:"
    [ ! -z "$REMAINING_GO" ] && echo "$REMAINING_GO"
    [ ! -z "$REMAINING_DEBUG" ] && echo "$REMAINING_DEBUG"
    [ ! -z "$REMAINING_SERVER" ] && echo "$REMAINING_SERVER"
fi

echo ""
echo "💀 Limpieza extrema completada."
echo "💡 Ahora puedes reiniciar VS Code si es necesario."
