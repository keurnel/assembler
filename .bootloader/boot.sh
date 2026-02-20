#!/usr/bin/env bash
# =============================================================================
# boot.sh - Bouw en boot de x86_64 bootloader in QEMU
# =============================================================================
# Gebruik: ./boot.sh
# =============================================================================

set -euo pipefail

SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
ASM_FILE="${SCRIPT_DIR}/bootloader.asm"
BIN_FILE="${SCRIPT_DIR}/bootloader.bin"

# Kleuren voor output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo -e "${YELLOW}=== Keurnel Bootloader Builder ===${NC}"
echo ""

# Controleer of nasm geïnstalleerd is
if ! command -v nasm &> /dev/null; then
    echo -e "${RED}[FOUT] nasm is niet geïnstalleerd.${NC}"
    echo "Installeer met:"
    echo "  Ubuntu/Debian: sudo apt install nasm"
    echo "  Fedora:        sudo dnf install nasm"
    echo "  Arch:          sudo pacman -S nasm"
    exit 1
fi

# Controleer of qemu geïnstalleerd is
if ! command -v qemu-system-x86_64 &> /dev/null; then
    echo -e "${RED}[FOUT] qemu-system-x86_64 is niet geïnstalleerd.${NC}"
    echo "Installeer met:"
    echo "  Ubuntu/Debian: sudo apt install qemu-system-x86"
    echo "  Fedora:        sudo dnf install qemu-system-x86"
    echo "  Arch:          sudo pacman -S qemu-system-x86"
    exit 1
fi

# Assembleer de bootloader
echo -e "${GREEN}[1/2] Assembleren van bootloader...${NC}"
nasm -f bin "${ASM_FILE}" -o "${BIN_FILE}"
echo "      Binair bestand: ${BIN_FILE} ($(stat -c%s "${BIN_FILE}") bytes)"

# Boot in QEMU
echo -e "${GREEN}[2/2] Booten in QEMU...${NC}"
echo ""
qemu-system-x86_64 \
    -drive format=raw,file="${BIN_FILE}" \
    -m 128M \
    -display curses \
    2>/dev/null

echo ""
echo -e "${GREEN}QEMU afgesloten.${NC}"

