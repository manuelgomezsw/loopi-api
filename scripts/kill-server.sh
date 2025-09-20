#!/bin/bash

# Script para limpiar procesos del servidor que puedan estar ocupando el puerto
# Útil cuando el IDE no puede iniciar el servidor por conflictos de puerto

echo "🧹 Limpiando procesos del servidor..."

# Función para mostrar procesos que usan el puerto 8080
show_port_usage() {
    echo "📋 Procesos usando el puerto 8080:"
    lsof -i :8080 2>/dev/null || echo "   (ninguno)"
}

# Mostrar estado inicial
show_port_usage

# 1. Matar procesos Go que ejecuten el servidor
echo ""
echo "🔍 Buscando procesos 'go run' del servidor..."
PROCESSES=$(pgrep -f "go run.*server" 2>/dev/null || true)
if [ ! -z "$PROCESSES" ]; then
    echo "🔪 Matando procesos 'go run': $PROCESSES"
    pkill -f "go run.*server" 2>/dev/null || true
    sleep 1
else
    echo "✅ No se encontraron procesos 'go run' del servidor"
fi

# 2. Matar procesos del debugger de VS Code (__debug_b*)
echo ""
echo "🔍 Buscando procesos del debugger de VS Code..."
DEBUG_PROCESSES=$(pgrep -f "__debug_b" 2>/dev/null || true)
if [ ! -z "$DEBUG_PROCESSES" ]; then
    echo "🔪 Matando procesos del debugger: $DEBUG_PROCESSES"
    pkill -f "__debug_b" 2>/dev/null || true
    sleep 1
else
    echo "✅ No se encontraron procesos del debugger"
fi

# 3. Matar procesos compilados del servidor (./server, server)
echo ""
echo "🔍 Buscando binarios del servidor..."
SERVER_PROCESSES=$(pgrep -f "server$|./server" 2>/dev/null || true)
if [ ! -z "$SERVER_PROCESSES" ]; then
    echo "🔪 Matando binarios del servidor: $SERVER_PROCESSES"
    pkill -f "server$|./server" 2>/dev/null || true
    sleep 1
else
    echo "✅ No se encontraron binarios del servidor"
fi

# 4. Matar cualquier proceso que use el puerto 8080 (fuerza bruta)
echo ""
echo "🔍 Buscando cualquier proceso en puerto 8080..."
PORT_PROCESSES=$(lsof -ti :8080 2>/dev/null || true)
if [ ! -z "$PORT_PROCESSES" ]; then
    echo "🔪 Liberando puerto 8080, matando PIDs: $PORT_PROCESSES"
    # Primero intentar con TERM
    echo "$PORT_PROCESSES" | xargs kill 2>/dev/null || true
    sleep 2
    # Si aún hay procesos, usar KILL
    REMAINING=$(lsof -ti :8080 2>/dev/null || true)
    if [ ! -z "$REMAINING" ]; then
        echo "🔪 Forzando cierre con SIGKILL: $REMAINING"
        echo "$REMAINING" | xargs kill -9 2>/dev/null || true
        sleep 1
    fi
else
    echo "✅ Puerto 8080 ya está libre"
fi

# 5. Verificación final
echo ""
echo "🔍 Verificación final del puerto 8080..."
if lsof -i :8080 >/dev/null 2>&1; then
    echo "❌ El puerto 8080 aún está ocupado:"
    lsof -i :8080
    echo ""
    echo "💡 Puedes intentar matar manualmente con:"
    lsof -ti :8080 | while read pid; do
        echo "   kill -9 $pid"
    done
    exit 1
else
    echo "✅ Puerto 8080 liberado exitosamente"
fi

echo ""
echo "🎉 Limpieza completada. Ahora puedes iniciar el servidor desde el IDE."
