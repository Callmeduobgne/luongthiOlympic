#!/bin/bash

# Copyright 2025 IBN Network (ICTU Blockchain Network)
#
# Licensed under the Apache License, Version 2.0 (the "License");
# you may not use this file except in compliance with the License.
# You may obtain a copy of the License at
#
#     http://www.apache.org/licenses/LICENSE-2.0
#
# Unless required by applicable law or agreed to in writing, software
# distributed under the License is distributed on an "AS IS" BASIS,
# WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
# See the License for the specific language governing permissions and
# limitations under the License.

# IBN Network - Dependency Installation Script
# This script installs all required technologies for the IBN Network project
# Compatible with: Ubuntu/Debian, WSL, macOS (with Homebrew)

set -e  # Exit on error

# Colors and Formatting
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
BOLD='\033[1m'
NC='\033[0m' # No Color

# UI Constants
WIDTH=60
DIVIDER="────────────────────────────────────────────────────────────"

# Functions
print_banner() {
    clear
    echo -e "${CYAN}${BOLD}"
    echo "    ██╗██████╗ ███╗   ██╗    ███████╗███████╗████████╗██╗   ██╗██████╗ "
    echo "    ██║██╔══██╗████╗  ██║    ██╔════╝██╔════╝╚══██╔══╝██║   ██║██╔══██╗"
    echo "    ██║██████╔╝██╔██╗ ██║    ███████╗█████╗     ██║   ██║   ██║██████╔╝"
    echo "    ██║██╔══██╗██║╚██╗██║    ╚════██║██╔══╝     ██║   ██║   ██║██╔═══╝ "
    echo "    ██║██████╔╝██║ ╚████║    ███████║███████╗   ██║   ╚██████╔╝██║     "
    echo "    ╚═╝╚═════╝ ╚═╝  ╚═══╝    ╚══════╝╚══════╝   ╚═╝    ╚═════╝ ╚═╝     "
    echo -e "${NC}"
    echo -e "${BLUE}${BOLD}                 IBN Network Setup Assistant                  ${NC}"
    echo -e "${BLUE}                 ===========================                  ${NC}"
    echo ""
}

print_header() {
    local text="$1"
    local text_len=${#text}
    local padding=$((58 - text_len))
    if [ $padding -lt 0 ]; then padding=0; fi
    local spaces=$(printf "%${padding}s")
    
    echo ""
    echo -e "${CYAN}${BOLD}╔${DIVIDER}╗${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${BOLD}${text}${NC}${spaces}${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}╚${DIVIDER}╝${NC}"
    echo ""
}

get_timestamp() {
    date "+%H:%M:%S"
}

print_success() {
    echo -e "${GREEN}[SUCCESS] ${NC}$1"
}

print_error() {
    echo -e "${RED}[ERROR]   ${NC}$1"
}

print_warning() {
    echo -e "${YELLOW}[WARNING] ${NC}$1"
}

print_info() {
    echo -e "${BLUE}[INFO]    ${NC}$1"
}

print_step() {
    echo -e "${CYAN}[STEP]    ${NC}$1"
}

check_command() {
    if command -v "$1" &> /dev/null; then
        return 0
    else
        return 1
    fi
}

# Compare two version numbers
# Returns: 0 if v1 == v2, 1 if v1 > v2, 2 if v1 < v2
compare_versions() {
    local v1=$1
    local v2=$2
    
    # Remove 'v' prefix if present
    v1=${v1#v}
    v2=${v2#v}
    
    # Split version into parts
    IFS='.' read -ra v1_parts <<< "$v1"
    IFS='.' read -ra v2_parts <<< "$v2"
    
    # Compare each part
    local max_len=${#v1_parts[@]}
    if [ ${#v2_parts[@]} -gt $max_len ]; then
        max_len=${#v2_parts[@]}
    fi
    
    for ((i=0; i<max_len; i++)); do
        local v1_part=${v1_parts[i]:-0}
        local v2_part=${v2_parts[i]:-0}
        
        # Remove non-numeric characters
        v1_part=$(echo "$v1_part" | sed 's/[^0-9]//g')
        v2_part=$(echo "$v2_part" | sed 's/[^0-9]//g')
        
        if [ -z "$v1_part" ]; then v1_part=0; fi
        if [ -z "$v2_part" ]; then v2_part=0; fi
        
        if [ "$v1_part" -gt "$v2_part" ]; then
            return 1  # v1 > v2
        elif [ "$v1_part" -lt "$v2_part" ]; then
            return 2  # v1 < v2
        fi
    done
    
    return 0  # v1 == v2
}

# Check if version meets minimum requirement
version_meets_requirement() {
    local installed_version=$1
    local required_version=$2
    
    compare_versions "$installed_version" "$required_version"
    local result=$?
    
    # result: 0 = equal, 1 = installed > required, 2 = installed < required
    if [ $result -eq 0 ] || [ $result -eq 1 ]; then
        return 0  # Meets requirement
    else
        return 1  # Does not meet requirement
    fi
}

# Get installed version of a command
get_installed_version() {
    local cmd=$1
    local version_pattern=$2
    
    if ! check_command "$cmd"; then
        echo ""
        return 1
    fi
    
    case "$cmd" in
        go)
            go version | awk '{print $3}' | sed 's/go//'
            ;;
        node)
            node --version | sed 's/v//'
            ;;
        npm)
            npm --version
            ;;
        docker)
            docker --version | awk '{print $3}' | sed 's/,//' | sed 's/v//'
            ;;
        git)
            git --version | awk '{print $3}'
            ;;
        sqlc)
            sqlc version 2>/dev/null | head -n1 | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -n1 || echo ""
            ;;
        air)
            air -v 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -n1 || echo ""
            ;;
        golangci-lint)
            golangci-lint version 2>/dev/null | grep -oE 'v?[0-9]+\.[0-9]+\.[0-9]+' | head -n1 || echo ""
            ;;
        *)
            echo ""
            ;;
    esac
}

# Check if tool needs installation or update
check_tool_status() {
    local tool_name=$1
    local required_version=$2
    local min_version=$3
    
    if [ -z "$min_version" ]; then
        min_version=$required_version
    fi
    
    if ! check_command "$tool_name"; then
        echo "missing"  # Tool not installed
        return
    fi
    
    local installed_version=$(get_installed_version "$tool_name")
    
    if [ -z "$installed_version" ]; then
        echo "unknown"  # Cannot determine version
        return
    fi
    
    # Check if meets minimum requirement
    if version_meets_requirement "$installed_version" "$min_version"; then
        echo "ok:$installed_version"  # Version is OK
    else
        echo "outdated:$installed_version"  # Version is outdated
    fi
}

# Ask user about usage mode
ask_usage_mode() {
    echo ""
    echo -e "${CYAN}${BOLD}╔${DIVIDER}╗${NC}"
    echo -e "${CYAN}${BOLD}║${NC}     ${YELLOW}${BOLD}BẠN MUỐN SỬ DỤNG DỰ ÁN NHƯ THẾ NÀO?${NC}                    ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}╠${DIVIDER}╣${NC}"
    echo -e "${CYAN}${BOLD}║${NC}                                                            ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[0] Production (Docker-only)${NC}                              ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}      ${BLUE}→ Chỉ cần Docker, không cần cài Go/Node.js${NC}            ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}      ${BLUE}→ Phù hợp cho: Deploy, test production${NC}                ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}                                                            ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[1] Development${NC}                                           ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}      ${BLUE}→ Cần tất cả công nghệ: Go, Node.js, Docker${NC}           ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}      ${BLUE}→ Phù hợp cho: Code, test, build, debug${NC}               ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}                                                            ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}╚${DIVIDER}╝${NC}"
    echo ""
    echo -ne "${BOLD}➜ Nhập đáp án của bạn (0-1) [mặc định: 1]: ${NC}"
    read usage_choice
    
    case $usage_choice in
        0|"docker-only")
            USAGE_MODE="docker-only"
            print_success "Đã chọn: Chế độ Docker-only (Production)"
            ;;
        1|""|"dev"|"development")
            USAGE_MODE="development"
            print_success "Đã chọn: Chế độ Development"
            ;;
        *)
            print_warning "Lựa chọn không hợp lệ. Sử dụng chế độ Development."
            USAGE_MODE="development"
            ;;
    esac
    echo ""
}

# Note: Script always uses fixed versions (UPDATE_MODE="required") for project compatibility
# All tools will be installed with exact versions specified in version requirements above

get_os() {
    if [[ "$OSTYPE" == "linux-gnu"* ]]; then
        if grep -q Microsoft /proc/version; then
            echo "wsl"
        else
            echo "linux"
        fi
    elif [[ "$OSTYPE" == "darwin"* ]]; then
        echo "macos"
    elif [[ "$OSTYPE" == "msys" ]] || [[ "$OSTYPE" == "cygwin" ]] || [[ -n "$WINDIR" ]]; then
        echo "windows"
    else
        echo "unknown"
    fi
}

ask_os() {
    echo ""
    echo -e "${CYAN}${BOLD}╔${DIVIDER}╗${NC}"
    echo -e "${CYAN}${BOLD}║${NC}     ${YELLOW}${BOLD}CHỌN HỆ ĐIỀU HÀNH CỦA BẠN${NC}                              ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}╠${DIVIDER}╣${NC}"
    echo -e "${CYAN}${BOLD}║${NC}                                                            ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[0] Windows${NC}                                               ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[1] Linux (Ubuntu/Debian)${NC}                                 ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[2] macOS${NC}                                                 ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[3] WSL (Windows Subsystem for Linux)${NC}                     ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}  ${GREEN}${BOLD}[4] Other/Unknown${NC}                                         ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}║${NC}                                                            ${CYAN}${BOLD}║${NC}"
    echo -e "${CYAN}${BOLD}╚${DIVIDER}╝${NC}"
    echo ""
    echo -ne "${BOLD}➜ Nhập đáp án của bạn (0-4): ${NC}"
    read os_choice
    
    case $os_choice in
        0)
            echo "windows"
            ;;
        1)
            echo "linux"
            ;;
        2)
            echo "macos"
            ;;
        3)
            echo "wsl"
            ;;
        4)
            echo "unknown"
            ;;
        *)
            print_error "Lựa chọn không hợp lệ. Mặc định: Linux"
            echo "linux"
            ;;
    esac
}

ask_windows_shell() {
    echo ""
    echo -e "${BLUE}╔════════════════════════════════════════════════════════╗${NC}"
    echo -e "${BLUE}║${NC}     ${YELLOW}Bạn Đang Sử Dụng Shell Nào Trên Windows?${NC}        ${BLUE}║${NC}"
    echo -e "${BLUE}╠════════════════════════════════════════════════════════╣${NC}"
    echo -e "${BLUE}║${NC}                                                      ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}  ${GREEN}(0)${NC}  PowerShell                                    ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}  ${GREEN}(1)${NC}  Command Prompt (CMD)                         ${BLUE}║${NC}"
    echo -e "${BLUE}║${NC}                                                      ${BLUE}║${NC}"
    echo -e "${BLUE}╚════════════════════════════════════════════════════════╝${NC}"
    echo ""
    echo -e "${YELLOW}┌──────────────────────────────────────────────────────┐${NC}"
    echo -e "${YELLOW}│${NC}  ${BLUE}Nhập đáp án của bạn (0-1):${NC}                          ${YELLOW}│${NC}"
    echo -e "${YELLOW}└──────────────────────────────────────────────────────┘${NC}"
    echo -ne "${GREEN}➜${NC} "
    read shell_choice
    
    case $shell_choice in
        0)
            echo "powershell"
            ;;
        1)
            echo "cmd"
            ;;
        *)
            print_warning "Lựa chọn không hợp lệ. Mặc định: PowerShell"
            echo "powershell"
            ;;
    esac
}

handle_windows() {
    print_header "Windows Installation Guide"
    echo ""
    print_warning "This script is designed for Linux/macOS/WSL."
    print_info "For Windows, please use the following options:"
    echo ""
    
    WINDOWS_SHELL=$(ask_windows_shell)
    
    if [[ "$WINDOWS_SHELL" == "powershell" ]]; then
        print_info "PowerShell script will be generated for you."
        generate_windows_powershell_script
    else
        print_info "Command Prompt installation guide:"
        print_windows_cmd_guide
    fi
    
    echo ""
    print_info "Alternatively, you can:"
    echo "  1. Use WSL (Windows Subsystem for Linux) - Recommended"
    echo "  2. Install Docker Desktop for Windows"
    echo "  3. Use Windows-native installers for each tool"
    echo ""
    read -p "Do you want to continue with Windows installation guide? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 0
    fi
}

generate_windows_powershell_script() {
    print_info "Generating PowerShell installation script..."
    
    # Ensure scripts directory exists
    mkdir -p scripts
    
    cat > scripts/setup.ps1 << 'POWERSHELL_EOF'
# IBN Network - Dependency Installation Script for Windows PowerShell
# Run this script as Administrator: Right-click PowerShell -> Run as Administrator

$ErrorActionPreference = "Stop"

# Version requirements
$GO_VERSION = "1.24.6"
$NODE_VERSION = "20"  # LTS version

Write-Host "========================================" -ForegroundColor Cyan
Write-Host "IBN Network - Windows Installation" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Check if running as Administrator
$isAdmin = ([Security.Principal.WindowsPrincipal] [Security.Principal.WindowsIdentity]::GetCurrent()).IsInRole([Security.Principal.WindowsBuiltInRole]::Administrator)
if (-not $isAdmin) {
    Write-Host "⚠ Warning: This script should be run as Administrator" -ForegroundColor Yellow
    Write-Host "Some installations may require elevated privileges." -ForegroundColor Yellow
    Write-Host ""
}

# 1. Install Chocolatey (if not installed)
Write-Host "1. Checking Chocolatey..." -ForegroundColor Blue
if (Get-Command choco -ErrorAction SilentlyContinue) {
    Write-Host "✓ Chocolatey already installed" -ForegroundColor Green
} else {
    Write-Host "Installing Chocolatey..." -ForegroundColor Yellow
    Set-ExecutionPolicy Bypass -Scope Process -Force
    [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072
    iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))
    Write-Host "✓ Chocolatey installed successfully" -ForegroundColor Green
}
Write-Host ""

# 2. Install Git
Write-Host "2. Installing Git..." -ForegroundColor Blue
if (Get-Command git -ErrorAction SilentlyContinue) {
    $gitVersion = git --version
    Write-Host "✓ Git already installed: $gitVersion" -ForegroundColor Green
} else {
    choco install git -y
    Write-Host "✓ Git installed successfully" -ForegroundColor Green
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
}
Write-Host ""

# 3. Install Go
Write-Host "3. Installing Go $GO_VERSION..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    $goVersion = go version
    Write-Host "✓ Go already installed: $goVersion" -ForegroundColor Green
} else {
    choco install golang --version=$GO_VERSION -y
    Write-Host "✓ Go installed successfully" -ForegroundColor Green
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    $env:GOPATH = "$env:USERPROFILE\go"
    [System.Environment]::SetEnvironmentVariable("GOPATH", $env:GOPATH, "User")
    [System.Environment]::SetEnvironmentVariable("Path", $env:Path, "User")
}
Write-Host ""

# 4. Install Node.js
Write-Host "4. Installing Node.js LTS..." -ForegroundColor Blue
if (Get-Command node -ErrorAction SilentlyContinue) {
    $nodeVersion = node --version
    Write-Host "✓ Node.js already installed: $nodeVersion" -ForegroundColor Green
} else {
    choco install nodejs-lts --version=$NODE_VERSION -y
    Write-Host "✓ Node.js installed successfully" -ForegroundColor Green
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
}
Write-Host ""

# 5. Install Docker Desktop
Write-Host "5. Installing Docker Desktop..." -ForegroundColor Blue
if (Get-Command docker -ErrorAction SilentlyContinue) {
    $dockerVersion = docker --version
    Write-Host "✓ Docker already installed: $dockerVersion" -ForegroundColor Green
} else {
    Write-Host "Installing Docker Desktop..." -ForegroundColor Yellow
    choco install docker-desktop -y
    Write-Host "✓ Docker Desktop installed successfully" -ForegroundColor Green
    Write-Host "⚠ Please restart your computer after Docker Desktop installation" -ForegroundColor Yellow
}
Write-Host ""

# 6. Install Go tools (requires Go to be installed)
Write-Host "6. Installing Go Development Tools..." -ForegroundColor Blue
if (Get-Command go -ErrorAction SilentlyContinue) {
    # Refresh PATH
    $env:Path = [System.Environment]::GetEnvironmentVariable("Path","Machine") + ";" + [System.Environment]::GetEnvironmentVariable("Path","User")
    
    # Install sqlc
    if (-not (Get-Command sqlc -ErrorAction SilentlyContinue)) {
        Write-Host "Installing sqlc..." -ForegroundColor Yellow
        go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.26.0
        Write-Host "✓ sqlc installed" -ForegroundColor Green
    } else {
        Write-Host "✓ sqlc already installed" -ForegroundColor Green
    }
    
    # Install air
    if (-not (Get-Command air -ErrorAction SilentlyContinue)) {
        Write-Host "Installing air..." -ForegroundColor Yellow
        go install github.com/cosmtrek/air@v1.54.0
        Write-Host "✓ air installed" -ForegroundColor Green
    } else {
        Write-Host "✓ air already installed" -ForegroundColor Green
    }
    
    # Install golangci-lint
    if (-not (Get-Command golangci-lint -ErrorAction SilentlyContinue)) {
        Write-Host "Installing golangci-lint..." -ForegroundColor Yellow
        choco install golangci-lint -y
        Write-Host "✓ golangci-lint installed" -ForegroundColor Green
    } else {
        Write-Host "✓ golangci-lint already installed" -ForegroundColor Green
    }
} else {
    Write-Host "⚠ Go is not installed. Skipping Go tools installation." -ForegroundColor Yellow
}
Write-Host ""

# Summary
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Installation Summary" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

if (Get-Command go -ErrorAction SilentlyContinue) {
    Write-Host "  Go: $(go version)" -ForegroundColor Green
} else {
    Write-Host "  Go: Not installed" -ForegroundColor Red
}

if (Get-Command node -ErrorAction SilentlyContinue) {
    Write-Host "  Node.js: $(node --version)" -ForegroundColor Green
} else {
    Write-Host "  Node.js: Not installed" -ForegroundColor Red
}

if (Get-Command npm -ErrorAction SilentlyContinue) {
    Write-Host "  npm: $(npm --version)" -ForegroundColor Green
} else {
    Write-Host "  npm: Not installed" -ForegroundColor Red
}

if (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "  Docker: $(docker --version)" -ForegroundColor Green
} else {
    Write-Host "  Docker: Not installed" -ForegroundColor Red
}

if (Get-Command git -ErrorAction SilentlyContinue) {
    Write-Host "  Git: $(git --version)" -ForegroundColor Green
} else {
    Write-Host "  Git: Not installed" -ForegroundColor Red
}

Write-Host ""
Write-Host "========================================" -ForegroundColor Cyan
Write-Host "Installation Complete!" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Blue
Write-Host "  1. Restart your terminal/PowerShell"
Write-Host "  2. Navigate to project: cd backend"
Write-Host "  3. Install Go dependencies: go mod download"
Write-Host "  4. Install Node.js dependencies: cd ..\frontend && npm install"
Write-Host "  5. Install Chaincode dependencies: cd ..\teaTraceCC && npm install"
Write-Host "  6. Start Docker Desktop and run: docker compose up -d"
Write-Host ""
POWERSHELL_EOF
    
    print_success "PowerShell script generated: scripts/setup.ps1"
    print_info "To run the PowerShell script:"
    echo "  1. Right-click PowerShell"
    echo "  2. Select 'Run as Administrator'"
    echo "  3. Navigate to project: cd E:\luongbeo"
    echo "  4. Run: .\scripts\setup.ps1"
    echo ""
}

print_windows_cmd_guide() {
    echo ""
    print_info "Command Prompt Installation Steps:"
    echo ""
    echo "1. Install Chocolatey (Package Manager):"
    echo "   - Open PowerShell as Administrator"
    echo "   - Run: Set-ExecutionPolicy Bypass -Scope Process -Force; [System.Net.ServicePointManager]::SecurityProtocol = [System.Net.ServicePointManager]::SecurityProtocol -bor 3072; iex ((New-Object System.Net.WebClient).DownloadString('https://community.chocolatey.org/install.ps1'))"
    echo ""
    echo "2. Install Git:"
    echo "   choco install git -y"
    echo ""
    echo "3. Install Go:"
    echo "   choco install golang --version=1.24.6 -y"
    echo ""
    echo "4. Install Node.js:"
    echo "   choco install nodejs-lts --version=20 -y"
    echo ""
    echo "5. Install Docker Desktop:"
    echo "   choco install docker-desktop -y"
    echo ""
    echo "6. Install Go tools (after Go is installed):"
    echo "   go install github.com/sqlc-dev/sqlc/cmd/sqlc@v1.26.0"
    echo "   go install github.com/cosmtrek/air@v1.54.0"
    echo "   choco install golangci-lint -y"
    echo ""
    echo "7. Restart your terminal after installation"
    echo ""
}

# Create NodeOUs config.yaml for MSPs
create_nodeou_configs() {
    local orgs_dir="$1"
    
    print_info "Đang tạo NodeOUs config..."
    
    # Tạo config cho Peer Organization
    local peer_msp="$orgs_dir/peerOrganizations/org1.ibn.vn/msp"
    if [ -d "$peer_msp" ]; then
        cat > "$peer_msp/config.yaml" << EOF
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: cacerts/ca.org1.ibn.vn-cert.pem
    OrganizationalUnitIdentifier: orderer
EOF
        print_success "Đã tạo config cho Org1MSP"
    fi
    
    # Tạo config cho Orderer Organization
    local orderer_msp="$orgs_dir/ordererOrganizations/ibn.vn/msp"
    if [ -d "$orderer_msp" ]; then
        # Note: Orderer CA cert name might vary, checking pattern
        local ca_cert="cacerts/ca.ibn.vn-cert.pem"
        # If specific file doesn't exist, try to find it
        if [ ! -f "$orderer_msp/$ca_cert" ]; then
             local found_cert=$(ls "$orderer_msp/cacerts/" | head -1)
             if [ -n "$found_cert" ]; then
                 ca_cert="cacerts/$found_cert"
             fi
        fi

        cat > "$orderer_msp/config.yaml" << EOF
NodeOUs:
  Enable: true
  ClientOUIdentifier:
    Certificate: $ca_cert
    OrganizationalUnitIdentifier: client
  PeerOUIdentifier:
    Certificate: $ca_cert
    OrganizationalUnitIdentifier: peer
  AdminOUIdentifier:
    Certificate: $ca_cert
    OrganizationalUnitIdentifier: admin
  OrdererOUIdentifier:
    Certificate: $ca_cert
    OrganizationalUnitIdentifier: orderer
EOF
        print_success "Đã tạo config cho OrdererMSP"
    fi
}

# Fix TLS CA cross-reference for orderers and peers to trust each other
fix_tls_ca_references() {
    print_info "Đang cấu hình TLS CA cross-references..."
    
    local orgs_dir="$1"
    local orderer_tls_ca="${orgs_dir}/ordererOrganizations/ibn.vn/tlsca/tlsca.ibn.vn-cert.pem"
    local peer_tls_ca="${orgs_dir}/peerOrganizations/org1.ibn.vn/tlsca/tlsca.org1.ibn.vn-cert.pem"
    
    # Check if TLS CA certs exist
    if [ ! -f "$orderer_tls_ca" ]; then
        print_warning "Orderer TLS CA not found: $orderer_tls_ca"
        return 1
    fi
    
    if [ ! -f "$peer_tls_ca" ]; then
        print_warning "Peer TLS CA not found: $peer_tls_ca"
        return 1
    fi
    
    # Copy Orderer TLS CA to Peer MSP tlscacerts (so peers trust orderers)
    print_info "Copying Orderer TLS CA to Peer MSP..."
    local peer_msp_tlsca="${orgs_dir}/peerOrganizations/org1.ibn.vn/msp/tlscacerts"
    mkdir -p "$peer_msp_tlsca"
    cp "$orderer_tls_ca" "$peer_msp_tlsca/tlsca.ibn.vn-cert.pem"
    
    # Copy Peer TLS CA to Orderer MSP tlscacerts (so orderers trust peers)
    print_info "Copying Peer TLS CA to Orderer MSP..."
    local orderer_msp_tlsca="${orgs_dir}/ordererOrganizations/ibn.vn/msp/tlscacerts"
    mkdir -p "$orderer_msp_tlsca"
    cp "$peer_tls_ca" "$orderer_msp_tlsca/tlsca.org1.ibn.vn-cert.pem"
    
    # Also copy to each peer's MSP tlscacerts
    for peer in peer0 peer1 peer2; do
        local peer_node_msp="${orgs_dir}/peerOrganizations/org1.ibn.vn/peers/${peer}.org1.ibn.vn/msp/tlscacerts"
        if [ -d "${orgs_dir}/peerOrganizations/org1.ibn.vn/peers/${peer}.org1.ibn.vn" ]; then
            mkdir -p "$peer_node_msp"
            cp "$orderer_tls_ca" "$peer_node_msp/tlsca.ibn.vn-cert.pem"
            print_success "Updated TLS CA for ${peer}.org1.ibn.vn"
        fi
    done
    
    # Copy to each orderer's MSP tlscacerts
    for orderer in orderer orderer1 orderer2; do
        local orderer_node_msp="${orgs_dir}/ordererOrganizations/ibn.vn/orderers/${orderer}.ibn.vn/msp/tlscacerts"
        if [ -d "${orgs_dir}/ordererOrganizations/ibn.vn/orderers/${orderer}.ibn.vn" ]; then
            mkdir -p "$orderer_node_msp"
            cp "$peer_tls_ca" "$orderer_node_msp/tlsca.org1.ibn.vn-cert.pem"
            print_success "Updated TLS CA for ${orderer}.ibn.vn"
        fi
    done
    
    print_success "TLS CA cross-references configured successfully"
    return 0
}

# Verify TLS certificates consistency
verify_tls_certificates() {
    print_info "Đang verify TLS certificates..."
    
    local orgs_dir="$1"
    local errors=0
    
    # Verify orderer certificates
    for orderer in orderer orderer1 orderer2; do
        local cert="${orgs_dir}/ordererOrganizations/ibn.vn/orderers/${orderer}.ibn.vn/tls/server.crt"
        if [ ! -f "$cert" ]; then
            print_error "Missing TLS cert for ${orderer}.ibn.vn"
            errors=$((errors + 1))
        else
            # Check if cert has correct SAN (Subject Alternative Name)
            if command -v openssl &> /dev/null; then
                local sans=$(openssl x509 -in "$cert" -noout -text 2>/dev/null | grep -A1 "Subject Alternative Name" | tail -1)
                if echo "$sans" | grep -q "${orderer}.ibn.vn"; then
                    print_success "✓ ${orderer}.ibn.vn certificate valid"
                else
                    print_warning "⚠ ${orderer}.ibn.vn certificate may have hostname mismatch"
                fi
            fi
        fi
    done
    
    # Verify peer certificates
    for peer in peer0 peer1 peer2; do
        local cert="${orgs_dir}/peerOrganizations/org1.ibn.vn/peers/${peer}.org1.ibn.vn/tls/server.crt"
        if [ ! -f "$cert" ]; then
            print_error "Missing TLS cert for ${peer}.org1.ibn.vn"
            errors=$((errors + 1))
        else
            if command -v openssl &> /dev/null; then
                local sans=$(openssl x509 -in "$cert" -noout -text 2>/dev/null | grep -A1 "Subject Alternative Name" | tail -1)
                if echo "$sans" | grep -q "${peer}.org1.ibn.vn"; then
                    print_success "✓ ${peer}.org1.ibn.vn certificate valid"
                else
                    print_warning "⚠ ${peer}.org1.ibn.vn certificate may have hostname mismatch"
                fi
            fi
        fi
    done
    
    if [ $errors -eq 0 ]; then
        print_success "✅ All TLS certificates verified"
        return 0
    else
        print_error "❌ Found $errors TLS certificate errors"
        return 1
    fi
}

# Wait for Fabric network to be fully ready after container start
wait_for_fabric_network() {
    print_info "Đang chờ Fabric network sẵn sàng..."
    local max_wait=180  # 3 minutes max
    local elapsed=0
    local check_interval=10
    
    # Wait for orderers to establish Raft cluster
    print_info "Checking orderer Raft cluster formation..."
    while [ $elapsed -lt $max_wait ]; do
        # Check if orderer logs show cluster is ready
        local orderer_ready=$(docker logs orderer.ibn.vn 2>&1 | grep -c "Raft leader changed.*channel=system-channel" 2>/dev/null | head -1)
        orderer_ready=${orderer_ready:-0}  # Default to 0 if empty
        if [ "$orderer_ready" -gt 0 ] 2>/dev/null; then
            print_success "✓ Orderer Raft cluster formed"
            break
        fi
        
        echo -n "."
        sleep $check_interval
        elapsed=$((elapsed + check_interval))
    done
    echo ""
    
    if [ $elapsed -ge $max_wait ]; then
        print_warning "Orderer cluster formation timeout, but continuing..."
    fi
    
    # Wait for peers to connect via gossip
    print_info "Checking peer gossip connections..."
    elapsed=0
    while [ $elapsed -lt $max_wait ]; do
        local peer_connected=$(docker logs peer0.org1.ibn.vn 2>&1 | grep -c "Membership view has changed" 2>/dev/null | head -1)
        peer_connected=${peer_connected:-0}  # Default to 0 if empty
        if [ "$peer_connected" -gt 0 ] 2>/dev/null; then
            print_success "✓ Peer gossip network established"
            break
        fi
        
        echo -n "."
        sleep $check_interval
        elapsed=$((elapsed + check_interval))
    done
    echo ""
    
    if [ $elapsed -ge $max_wait ]; then
        print_warning "Peer gossip timeout, but continuing..."
    fi
    
    # Final check: No critical TLS errors in last 30 seconds of logs
    print_info "Checking for TLS errors..."
    local tls_errors=0
    for container in orderer.ibn.vn orderer1.ibn.vn orderer2.ibn.vn peer0.org1.ibn.vn peer1.org1.ibn.vn peer2.org1.ibn.vn; do
        # Check if container is running first
        if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
            print_warning "Container $container is not running, skipping TLS check"
            continue
        fi
        
        local recent_tls_errors=$(docker logs --since 30s "$container" 2>&1 | grep -c "tls: failed to verify certificate" 2>/dev/null | head -1)
        recent_tls_errors=${recent_tls_errors:-0}  # Default to 0 if empty
        
        if [ "$recent_tls_errors" -gt 5 ] 2>/dev/null; then
            print_error "⚠ Container $container có nhiều TLS errors: $recent_tls_errors"
            tls_errors=$((tls_errors + 1))
        fi
    done
    
    if [ $tls_errors -eq 0 ]; then
        print_success "✅ No critical TLS errors detected"
        return 0
    else
        print_error "❌ Detected TLS errors in $tls_errors containers"
        print_warning "Network may not be stable. Check logs with option 7."
        return 1
    fi
}

# Generate crypto material using cryptogen
generate_crypto_material() {
    print_info "Đang tạo crypto material..."
    
    local core_dir="$1"
    local crypto_config="${core_dir}/crypto-config.yaml"
    local output_dir="${core_dir}/organizations"
    local genesis_dir="${core_dir}/system-genesis-block"
    local genesis_block="${genesis_dir}/genesis.block"
    
    if [ ! -f "$crypto_config" ]; then
        print_error "Không tìm thấy file config: $crypto_config"
        return 1
    fi
    
    # CRITICAL: Remove existing crypto material AND genesis block
    # Genesis block contains certificates, so it must be regenerated when crypto changes
    if [ -d "$output_dir" ]; then
        print_warning "Removing existing crypto material..."
        rm -rf "$output_dir"
    fi
    
    # Also remove genesis block if it exists (will be regenerated after crypto)
    if [ -f "$genesis_block" ]; then
        print_warning "Removing existing genesis block (will be regenerated with new certificates)..."
        rm -f "$genesis_block"
    fi
    
    # Try local cryptogen first
    if check_command cryptogen; then
        print_info "Sử dụng cryptogen local..."
        if cryptogen generate --config="$crypto_config" --output="$output_dir"; then
            print_success "Đã tạo crypto material thành công (local)"
            create_nodeou_configs "$output_dir"
            fix_tls_ca_references "$output_dir"
            verify_tls_certificates "$output_dir"
            return 0
        fi
    fi
    
    # Fallback to Docker
    if check_command docker; then
        print_info "Sử dụng Docker để chạy cryptogen..."
        # Get absolute path for volume mount
        local abs_core_dir="$(cd "$core_dir" && pwd)"
        
        if docker run --rm -v "$abs_core_dir":/core -w /core hyperledger/fabric-tools:2.5 \
            cryptogen generate --config=crypto-config.yaml --output=organizations; then
            print_success "Đã tạo crypto material thành công (Docker)"
            create_nodeou_configs "$output_dir"
            fix_tls_ca_references "$output_dir"
            verify_tls_certificates "$output_dir"
            return 0
        fi
    fi
    
    print_error "Không thể tạo crypto material. Vui lòng cài đặt cryptogen hoặc Docker."
    return 1
}

# Generate genesis block using configtxgen
generate_genesis_block() {
    print_info "Đang tạo genesis block..."
    
    local core_dir="$1"
    local genesis_dir="${core_dir}/system-genesis-block"
    local genesis_block="${genesis_dir}/genesis.block"
    # Profile name matches core/configtx/configtx.yaml
    local profile="RaftOrdererGenesis"
    local channel_id="system-channel"
    
    # CRITICAL: Remove old genesis block to prevent certificate mismatch
    if [ -f "$genesis_block" ]; then
        print_warning "Removing existing genesis block to prevent certificate mismatch..."
        rm -f "$genesis_block"
    fi
    
    mkdir -p "$genesis_dir"
    
    # Try local configtxgen first
    if check_command configtxgen; then
        print_info "Sử dụng configtxgen local..."
        # Point to directory containing configtx.yaml
        export FABRIC_CFG_PATH="${core_dir}/configtx"
        if configtxgen -profile "$profile" -channelID "$channel_id" -outputBlock "$genesis_block"; then
            print_success "Đã tạo genesis block thành công (local)"
            return 0
        fi
    fi
    
    # Fallback to Docker
    if check_command docker; then
        print_info "Sử dụng Docker để chạy configtxgen..."
        local abs_core_dir="$(cd "$core_dir" && pwd)"
        
        if docker run --rm -v "$abs_core_dir":/core -w /core \
            -e FABRIC_CFG_PATH=/core/configtx \
            hyperledger/fabric-tools:2.5 \
            configtxgen -profile "$profile" -channelID "$channel_id" -outputBlock "system-genesis-block/genesis.block"; then
            print_success "Đã tạo genesis block thành công (Docker)"
            return 0
        fi
    fi
    
    print_error "Không thể tạo genesis block. Vui lòng cài đặt configtxgen hoặc Docker."
    return 1
}

# Install project dependencies
install_project_dependencies() {
    print_header "Cài Đặt Thư Viện Dự Án"
    echo ""
    
    # Backend (Go)
    if [ -d "backend" ]; then
        print_info "Đang cài đặt dependencies cho Backend (Go)..."
        if check_command go; then
            (cd backend && go mod download)
            if [ $? -eq 0 ]; then
                print_success "Backend dependencies: OK"
            else
                print_warning "Không thể cài đặt Backend dependencies"
            fi
        else
            print_warning "Go chưa được cài đặt, bỏ qua Backend dependencies"
        fi
    fi
    echo ""
    
    # Frontend (Node.js)
    if [ -d "frontend" ]; then
        print_info "Đang cài đặt dependencies cho Frontend (Node.js)..."
        if check_command npm; then
            (cd frontend && npm install)
            if [ $? -eq 0 ]; then
                print_success "Frontend dependencies: OK"
            else
                print_warning "Không thể cài đặt Frontend dependencies"
            fi
        else
            print_warning "npm chưa được cài đặt, bỏ qua Frontend dependencies"
        fi
    fi
    echo ""
    
    # Chaincode (Node.js)
    if [ -d "teaTraceCC" ]; then
        print_info "Đang cài đặt dependencies cho Chaincode..."
        if check_command npm; then
            (cd teaTraceCC && npm install)
            if [ $? -eq 0 ]; then
                print_success "Chaincode dependencies: OK"
            else
                print_warning "Không thể cài đặt Chaincode dependencies"
            fi
        else
            print_warning "npm chưa được cài đặt, bỏ qua Chaincode dependencies"
        fi
    fi
    echo ""
}

# Check if container is running
check_fabric_container() {
    local container=$1
    if ! docker ps --format '{{.Names}}' | grep -q "^${container}$"; then
        print_error "Container ${container} is not running"
        return 1
    fi
    return 0
}

# Wait for container to be ready
wait_for_fabric_container() {
    local container=$1
    local max_attempts=30
    local attempt=0
    
    print_info "Waiting for container ${container} to be ready..."
    while [ $attempt -lt $max_attempts ]; do
        if docker exec "${container}" peer version > /dev/null 2>&1; then
            print_success "Container ${container} is ready"
            return 0
        fi
        attempt=$((attempt + 1))
        sleep 2
    done
    
    print_error "Container ${container} did not become ready in time"
    return 1
}

# Create channel - Professional implementation with validation
create_fabric_channel() {
    local channel_name=$1
    
    if [ -z "$channel_name" ]; then
        print_error "Channel name is required"
        return 1
    fi
    
    # Validate channel name format (alphanumeric, hyphens, underscores, lowercase)
    if ! echo "$channel_name" | grep -qE '^[a-z0-9_-]+$'; then
        print_error "Invalid channel name format. Use lowercase alphanumeric, hyphens, or underscores only."
        return 1
    fi
    
    print_header "Create Channel: ${channel_name}"
    echo ""
    
    # Check if orderer is running
    print_info "Checking orderer availability..."
    if ! check_fabric_container "orderer.ibn.vn"; then
        print_error "Orderer container is not running. Please start the network first."
        return 1
    fi
    print_success "Orderer is running"
    
    # Check if peer0 is running (we'll use it to create channel)
    local first_peer="peer0.org1.ibn.vn"
    print_info "Checking peer availability..."
    if ! check_fabric_container "${first_peer}"; then
        print_error "Peer container ${first_peer} is not running. Please start the network first."
        return 1
    fi
    print_success "Peer ${first_peer} is running"
    
    wait_for_fabric_container "${first_peer}"
    
    # Find project root to resolve paths correctly
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local project_root=""
    
    # Try to find project root from current directory
    if [ -f "docker-compose.yml" ]; then
        project_root="$(pwd)"
    elif [ -f "${script_dir}/../docker-compose.yml" ]; then
        project_root="$(cd "${script_dir}/.." && pwd)"
    elif [ -f "../docker-compose.yml" ]; then
        project_root="$(cd .. && pwd)"
    else
        # Search up to 3 levels
        local search_dir="$(pwd)"
        for i in {1..3}; do
            if [ -f "${search_dir}/docker-compose.yml" ]; then
                project_root="${search_dir}"
                break
            fi
            search_dir="$(dirname "$search_dir")"
        done
    fi
    
    if [ -z "$project_root" ]; then
        print_error "Cannot find project root directory (docker-compose.yml)"
        return 1
    fi
    
    # Channel transaction file path (use absolute path)
    local channel_tx="${project_root}/core/channel-artifacts/${channel_name}.tx"
    local channel_block="${project_root}/core/channel-artifacts/${channel_name}.block"
    
    # Create channel-artifacts directory if it doesn't exist
    mkdir -p "${project_root}/core/channel-artifacts"
    
    # Check if channel transaction file exists
    if [ ! -f "${channel_tx}" ]; then
        print_warning "Channel transaction file not found: ${channel_tx}"
        print_info "Attempting to generate channel transaction file..."
        
        # Determine profile name from configtx.yaml
        local profile_name="ThreePeersChannel"  # Default profile name from configtx.yaml
        
        # Try to detect profile name from configtx.yaml
        if [ -f "core/configtx/configtx.yaml" ]; then
            local detected_profile=$(grep -A 5 "Channel Profile" core/configtx/configtx.yaml | grep -E "^\s+[A-Za-z]+:" | head -1 | sed 's/.*: *//' | tr -d ' ')
            if [ -n "$detected_profile" ]; then
                profile_name="$detected_profile"
            fi
        fi
        
        print_info "Using profile: ${profile_name}"
        
        # Try to generate using configtxgen (local)
        if check_command configtxgen; then
            export FABRIC_CFG_PATH="core/configtx"
            if configtxgen -profile "${profile_name}" -outputCreateChannelTx "${channel_tx}" -channelID "${channel_name}" 2>/dev/null; then
                print_success "Generated channel transaction file using local configtxgen"
            else
                print_warning "Local configtxgen failed, trying Docker..."
                # Fall through to Docker method
            fi
        fi
        
        # Try Docker method if local failed or not available
        if [ ! -f "${channel_tx}" ]; then
            print_info "Using Docker to generate channel transaction file..."
            
            # Check if Docker is available
            if ! check_command docker; then
                print_error "Docker is not available. Cannot generate channel transaction file."
                return 1
            fi
            
            # Use project_root already found above
            local abs_core_dir="${project_root}/core"
            local abs_channel_artifacts="${abs_core_dir}/channel-artifacts"
            
            # Ensure channel-artifacts directory exists
            mkdir -p "${abs_channel_artifacts}"
            
            # Check if configtx.yaml exists
            if [ ! -f "${abs_core_dir}/configtx/configtx.yaml" ]; then
                print_error "configtx.yaml not found at: ${abs_core_dir}/configtx/configtx.yaml"
                print_info "Please ensure configtx configuration exists"
                return 1
            fi
            
            # Generate using Docker
            print_info "Running configtxgen in Docker container..."
            print_info "Profile: ${profile_name}, Channel: ${channel_name}"
            
            # In Fabric 2.x, we create a genesis block for the channel instead of a transaction file
            local channel_block="${abs_channel_artifacts}/${channel_name}.block"
            
            local docker_output=$(docker run --rm \
                -v "${abs_core_dir}:/core" \
                -w /core \
                -e FABRIC_CFG_PATH=/core/configtx \
                hyperledger/fabric-tools:2.5 \
                configtxgen \
                -profile "${profile_name}" \
                -outputBlock "channel-artifacts/${channel_name}.block" \
                -channelID "${channel_name}" 2>&1)
            
            local docker_status=$?
            
            if [ $docker_status -eq 0 ] && [ -f "${channel_block}" ]; then
                print_success "Generated channel genesis block using Docker"
                # Mark that we have genesis block for later use
                export CHANNEL_GENESIS_BLOCK="${channel_block}"
                # For compatibility, create a dummy .tx file pointing to the block
                echo "# Channel genesis block created at ${channel_block}" > "${channel_tx}"
            else
                print_error "Failed to generate channel genesis block"
                echo ""
                print_info "Docker output:"
                echo "$docker_output" | tail -20
                echo ""
                print_info "Troubleshooting:"
                echo "  1. Check if configtx.yaml exists at: core/configtx/configtx.yaml"
                echo "  2. Verify '${profile_name}' profile exists in configtx.yaml"
                echo "  3. Ensure Docker can access the core directory"
                echo "  4. Check Docker image: hyperledger/fabric-tools:2.5"
                return 1
            fi
        fi
    else
        print_success "Channel transaction file found: ${channel_tx}"
    fi
    
    # Check if channel already exists
    print_info "Checking if channel already exists..."
    local existing_check=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
        -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
        -w /root \
        "${first_peer}" \
        peer channel list 2>&1 | grep -i "${channel_name}" || echo "")
    
    if [ -n "$existing_check" ]; then
        print_warning "Channel ${channel_name} may already exist"
        echo -ne "${YELLOW}Do you want to continue anyway? (y/N): ${NC}"
        read -r response
        if [[ ! "$response" =~ ^[Yy]$ ]]; then
            print_info "Aborted by user"
            return 1
        fi
    fi
    
    # Copy channel transaction file to peer container
    print_info "Copying channel transaction file to ${first_peer}..."
    
    # Verify file exists before copying
    if [ ! -f "${channel_tx}" ]; then
        print_error "Channel transaction file not found: ${channel_tx}"
        print_info "Expected path: ${channel_tx}"
        return 1
    fi
    
    # Verify file is not empty
    if [ ! -s "${channel_tx}" ]; then
        print_error "Channel transaction file is empty: ${channel_tx}"
        return 1
    fi
    
    print_info "Source file: ${channel_tx}"
    # Copy to /root (working directory in container)
    if docker cp "${channel_tx}" "${first_peer}:/root/${channel_name}.tx" 2>&1; then
        print_success "Channel transaction file copied to /root/${channel_name}.tx"
    else
        print_error "Failed to copy channel transaction file"
        print_info "Troubleshooting:"
        echo "  1. Check if file exists: ls -la ${channel_tx}"
        echo "  2. Check file permissions"
        echo "  3. Check if peer container is accessible: docker ps | grep ${first_peer}"
        return 1
    fi
    
    # In Fabric 2.5.9, we don't use peer channel create with .tx file
    # Instead, we use the genesis block created by configtxgen
    print_info "Preparing channel ${channel_name} for peer joining..."
    echo ""
    
    # Check if we have genesis block
    local genesis_block="${project_root}/core/channel-artifacts/${channel_name}.block"
    if [ ! -f "${genesis_block}" ]; then
        print_error "Genesis block not found: ${genesis_block}"
        print_info "This should have been created by configtxgen"
        return 1
    fi
    
    # Copy genesis block to peer container
    print_info "Copying genesis block to peer container..."
    if docker cp "${genesis_block}" "${first_peer}:/root/${channel_name}.block" 2>&1; then
        print_success "Genesis block copied to peer container"
        export CHANNEL_BLOCK_AVAILABLE="true"
        export CHANNEL_JUST_CREATED="true"
        
        # Also copy to host channel-artifacts for consistency
        local channel_block="core/channel-artifacts/${channel_name}.block"
        print_success "Channel ${channel_name} is ready for peer joining"
        
        # Cleanup transaction file from container
        docker exec "${first_peer}" rm -f "/root/${channel_name}.tx" > /dev/null 2>&1
        
        echo ""
        print_success "Channel creation completed successfully!"
        return 0
    else
        print_error "Failed to copy genesis block to peer container"
        print_info "Troubleshooting:"
        echo "  1. Check if genesis block exists: ls -la ${genesis_block}"
        echo "  2. Check if peer container is running: docker ps | grep ${first_peer}"
        return 1
    fi
}

# Join peers to channel - Simplified production-ready implementation
join_fabric_peers() {
    local channel_name=$1
    local max_retries=${2:-3}
    
    if [ -z "$channel_name" ]; then
        print_error "Channel name is required"
        return 1
    fi
    
    print_header "Join Peers to Channel: ${channel_name}"
    echo ""
    
    # Define peer0 as the first peer
    local first_peer="peer0.org1.ibn.vn"
    local channel_block="core/channel-artifacts/${channel_name}.block"
    
    # Step 0: Copy Admin MSP to all peers (required for join operation)
    print_info "Step 0: Copying Admin MSP to peers..."
    local admin_msp_source="core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp"
    
    if [ ! -d "$admin_msp_source" ]; then
        print_error "Admin MSP not found: $admin_msp_source"
        return 1
    fi
    
    for peer in peer0.org1.ibn.vn peer1.org1.ibn.vn peer2.org1.ibn.vn; do
        if docker ps --format '{{.Names}}' | grep -q "^${peer}$"; then
            docker exec "$peer" rm -rf /tmp/admin_msp 2>/dev/null || true
            docker cp "$admin_msp_source" "${peer}:/tmp/admin_msp"
            print_success "✓ Admin MSP copied to $peer"
        fi
    done
    echo ""
    
    # Step 1: Ensure we have the block file on host
    print_info "Step 1: Verifying channel block file..."
    if [ ! -f "${channel_block}" ]; then
        print_warning "Block file not found on host: ${channel_block}"
        print_info "Attempting to retrieve from peer0 container..."
        
        # Try to copy from peer0 container
        if check_fabric_container "${first_peer}"; then
            if docker cp "${first_peer}:/root/${channel_name}.block" "${channel_block}" 2>/dev/null; then
                print_success "Block file retrieved from container"
            else
                print_error "Block file not found in container either"
                print_error "Please create the channel first (option 3)"
                return 1
            fi
        else
            print_error "Peer0 container is not running"
            return 1
        fi
    else
        print_success "Block file found: ${channel_block}"
    fi
    
    # Verify block file is not empty
    if [ ! -s "${channel_block}" ]; then
        print_error "Block file is empty: ${channel_block}"
        return 1
    fi
    
    echo ""
    print_info "Step 2: Joining peers to channel..."
    echo ""
    
    # Define all peers
    local peers=(
        "peer0.org1.ibn.vn:7051"
        "peer1.org1.ibn.vn:8051"
        "peer2.org1.ibn.vn:9051"
    )
    
    print_info "Network has ${#peers[@]} peers to join:"
    for i in "${!peers[@]}"; do
        echo "  $((i+1)). ${peers[$i]}"
    done
    echo ""
    
    local success_count=0
    local failed_peers=()
    
    # Join each peer
    for peer_info in "${peers[@]}"; do
        IFS=':' read -r peer_name peer_port <<< "${peer_info}"
        
        print_info "Processing ${peer_name}..."
        
        # Check if container is running
        if ! check_fabric_container "${peer_name}"; then
            print_warning "Container ${peer_name} is not running - skipping"
            failed_peers+=("${peer_name}")
            echo ""
            continue
        fi
        
        # Wait for container to be ready
        wait_for_fabric_container "${peer_name}"
        
        # Check if already joined
        print_info "Checking if ${peer_name} already joined channel..."
        local channel_list=$(docker exec \
            -e CORE_PEER_LOCALMSPID="Org1MSP" \
            -e CORE_PEER_TLS_ENABLED=true \
            -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
            -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
            -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
            "${peer_name}" \
            peer channel list 2>&1 | grep -i "${channel_name}" || echo "")
        
        if [ -n "$channel_list" ]; then
            print_success "${peer_name} already joined channel ${channel_name}"
            success_count=$((success_count + 1))
            echo ""
            continue
        fi
        
        # Copy block file to container
        print_info "Copying block file to ${peer_name}..."
        if ! docker cp "${channel_block}" "${peer_name}:/root/${channel_name}.block" 2>/dev/null; then
            print_error "Failed to copy block file to ${peer_name}"
            failed_peers+=("${peer_name}")
            echo ""
            continue
        fi
        
        # Verify block file in container
        if ! docker exec "${peer_name}" test -f "/root/${channel_name}.block" 2>/dev/null; then
            print_error "Block file not found in ${peer_name} container after copy"
            failed_peers+=("${peer_name}")
            echo ""
            continue
        fi
        
        print_success "Block file ready in ${peer_name}"
        
        # Join channel with retry logic
        print_info "Joining ${peer_name} to channel ${channel_name}..."
        local retry_count=0
        local join_success=false
        
        while [ $retry_count -lt $max_retries ]; do
            if [ $retry_count -gt 0 ]; then
                print_info "Retry attempt $((retry_count + 1))/${max_retries}..."
                sleep 2
            fi
            
            # Execute join command (use Admin MSP for proper authorization)
            local join_output=$(docker exec \
                -e CORE_PEER_LOCALMSPID="Org1MSP" \
                -e CORE_PEER_TLS_ENABLED=true \
                -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
                -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
                -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
                -w /root \
                "${peer_name}" \
                peer channel join -b "${channel_name}.block" 2>&1)
            
            local join_status=$?
            
            if [ $join_status -eq 0 ]; then
                print_success "${peer_name} joined channel ${channel_name}"
                # Show output for verification
                if [ -n "$join_output" ]; then
                    echo "$join_output" | grep -v "^$" | head -3
                fi
                success_count=$((success_count + 1))
                join_success=true
                break
            else
                retry_count=$((retry_count + 1))
                if [ $retry_count -lt $max_retries ]; then
                    print_warning "Join failed, retrying..."
                    echo "$join_output" | tail -3
                fi
            fi
        done
        
        if [ "$join_success" = false ]; then
            print_error "Failed to join ${peer_name} after ${max_retries} attempts"
            print_info "Error output:"
            echo "$join_output" | tail -5
            failed_peers+=("${peer_name}")
        fi
        
        echo ""
    done
    
    # Summary
    echo ""
    print_header "Join Summary"
    echo ""
    
    print_info "Total peers: ${#peers[@]}"
    print_info "Successfully joined: ${success_count}"
    print_info "Failed: ${#failed_peers[@]}"
    
    if [ ${#failed_peers[@]} -gt 0 ]; then
        print_warning "Failed peers: ${failed_peers[*]}"
    fi
    
    echo ""
    
    if [ $success_count -eq ${#peers[@]} ]; then
        print_success "All peers joined channel ${channel_name} successfully!"
        return 0
    elif [ $success_count -gt 0 ]; then
        print_warning "Partial success: ${success_count}/${#peers[@]} peers joined"
        return 1
    else
        print_error "No peers joined channel ${channel_name}"
        print_info "Please check:"
        echo "  1. Channel exists and was created successfully"
        echo "  2. Channel block file is valid"
        echo "  3. Peer containers are running and healthy"
        echo "  4. Network connectivity between peers and orderer"
        return 1
    fi
}


# Query installed chaincode to get package ID
query_installed_chaincode() {
    local peer_name=$1
    local chaincode_label=$2
    
    print_info "Querying installed chaincode on ${peer_name}..."
    
    # Use Admin MSP for lifecycle queries (required for ACL)
    local output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
        -e CORE_PEER_ADDRESS="${peer_name}:7051" \
        -w /root \
        "${peer_name}" \
        peer lifecycle chaincode queryinstalled 2>&1)
    
    if [ $? -ne 0 ]; then
        print_warning "Failed to query installed chaincode on ${peer_name} (may need Admin MSP)"
        # Try to extract from error message or output anyway
        echo "$output" | tail -10
    fi
    
    # Extract package ID from output - try multiple formats
    local package_id=""
    
    # Format 1: "Package ID: basic_1.0:abc123..., Label: basic_1.0"
    package_id=$(echo "$output" | grep -i "Package ID" | grep -i "${chaincode_label}" | head -1 | sed -n 's/.*Package ID: *\([^,]*\).*/\1/p' | tr -d ' ' | tr -d "'" | tr -d '"')
    
    # Format 2: Look for label:hash pattern (64+ hex chars)
    if [ -z "$package_id" ]; then
        package_id=$(echo "$output" | grep -i "${chaincode_label}" | grep -oE '[a-zA-Z0-9_]+:[a-f0-9]{64,}' | head -1)
    fi
    
    # Format 3: "Installed remotely: response:<status:200 payload:"\nEteaTraceCC_1.0:hash..."
    if [ -z "$package_id" ]; then
        package_id=$(echo "$output" | grep -oE "${chaincode_label}:[a-f0-9]{64,}" | head -1)
    fi
    
    # Format 4: "Installed chaincode with package ID 'teaTraceCC_1.0:hash...'"
    if [ -z "$package_id" ]; then
        package_id=$(echo "$output" | grep -oE "package ID ['\"]?${chaincode_label}:[a-f0-9]+" | sed "s/.*package ID ['\"]\?//i" | head -1)
    fi
    
    if [ -z "$package_id" ]; then
        print_warning "Could not extract package ID from output"
        # Don't return error, just warn - package ID might be extracted from install output
        return 1
    fi
    
    echo "$package_id"
    return 0
}

# Check commit readiness
check_commit_readiness() {
    local peer_name=$1
    local channel_name=$2
    local chaincode_name=$3
    local chaincode_version=$4
    local sequence=$5
    
    print_info "Checking commit readiness on ${peer_name}..."
    
    # Extract peer port from peer name
    local peer_port="7051"
    case "${peer_name}" in
        peer1.org1.ibn.vn) peer_port="8051" ;;
        peer2.org1.ibn.vn) peer_port="9051" ;;
        *) peer_port="7051" ;;
    esac
    
    # Use Admin MSP for lifecycle queries (required for ACL)
    local output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
        -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
        -w /root \
        "${peer_name}" \
        peer lifecycle chaincode checkcommitreadiness \
        --channelID "${channel_name}" \
        --name "${chaincode_name}" \
        --version "${chaincode_version}" \
        --sequence "${sequence}" \
        --tls \
        --cafile "/etc/hyperledger/fabric/tls/ca.crt" \
        --output json 2>&1)
    
    if [ $? -ne 0 ]; then
        print_warning "Failed to check commit readiness: $output"
        return 1
    fi
    
    # Check if all organizations are ready
    local ready=$(echo "$output" | grep -o '"approvals":{[^}]*}' | grep -o '"Org1MSP":true' || echo "")
    
    if [ -n "$ready" ]; then
        print_success "Chaincode is ready to commit"
        return 0
    else
        print_warning "Chaincode is not ready to commit yet"
        return 1
    fi
}

# Deploy chaincode - Full professional implementation
deploy_fabric_chaincode() {
    local channel_name=$1
    local chaincode_name=$2
    local chaincode_version=$3
    local chaincode_path=$4
    local sequence=${5:-1}  # Default sequence is 1
    local init_required=${6:-false}  # Default no init required
    
    if [ -z "$channel_name" ] || [ -z "$chaincode_name" ] || [ -z "$chaincode_version" ] || [ -z "$chaincode_path" ]; then
        print_error "Usage: deploy-chaincode <channel-name> <chaincode-name> <version> <path> [sequence] [init-required]"
        return 1
    fi
    
    print_header "Deploy Chaincode: ${chaincode_name} v${chaincode_version}"
    echo ""
    
    # Resolve absolute path for chaincode FIRST (before validation)
    local abs_chaincode_path
    if [[ "${chaincode_path}" == /* ]]; then
        abs_chaincode_path="${chaincode_path}"
    else
        # Get project root (where script is located)
        local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
        local project_root=""
        if [ -f "docker-compose.yml" ]; then
            project_root="$(pwd)"
        elif [ -f "${script_dir}/../docker-compose.yml" ]; then
            project_root="$(cd "${script_dir}/.." && pwd)"
        else
            project_root="$(pwd)"
        fi
        # Handle relative paths like "../teaTraceCC" or "teaTraceCC"
        if [[ "${chaincode_path}" == ../* ]]; then
            # Remove "../" prefix and resolve from project root parent
            local relative_part="${chaincode_path#../}"
            local parent_dir="$(cd "${project_root}/.." && pwd)"
            abs_chaincode_path="${parent_dir}/${relative_part}"
            # If not found, try from project root (in case path is wrong)
            if [ ! -d "${abs_chaincode_path}" ]; then
                abs_chaincode_path="${project_root}/${relative_part}"
            fi
        else
            abs_chaincode_path="${project_root}/${chaincode_path}"
        fi
    fi
    
    # Validate chaincode path (using resolved absolute path)
    if [ ! -d "${abs_chaincode_path}" ]; then
        print_error "Chaincode path not found: ${chaincode_path}"
        print_info "Resolved to: ${abs_chaincode_path}"
        print_info "Please provide absolute path or relative path from project root"
        return 1
    fi
    
    # Check if package.json exists (for Node.js chaincode)
    if [ ! -f "${abs_chaincode_path}/package.json" ]; then
        print_warning "package.json not found in ${abs_chaincode_path}"
        print_info "Assuming chaincode is already built or using different language"
    fi
    
    # Check if at least one peer has joined the channel (required for chaincode deployment)
    print_info "Verifying peers have joined channel ${channel_name}..."
    local peers=("peer0.org1.ibn.vn:7051" "peer1.org1.ibn.vn:8051" "peer2.org1.ibn.vn:9051")
    local joined_peers=0
    
    for peer_info in "${peers[@]}"; do
        IFS=':' read -r peer_name peer_port <<< "${peer_info}"
        if check_fabric_container "${peer_name}"; then
            local channel_list=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
                -e CORE_PEER_TLS_ENABLED=true \
                -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
                -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
                -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
                -w /root \
                "${peer_name}" \
                peer channel list 2>&1 | grep -i "${channel_name}" || echo "")
            
            if [ -n "$channel_list" ]; then
                joined_peers=$((joined_peers + 1))
            fi
        fi
    done
    
    if [ $joined_peers -eq 0 ]; then
        print_error "No peers have joined channel ${channel_name}"
        print_error "Cannot deploy chaincode without peers in channel"
        print_info "Please join peers to channel first (option 4)"
        return 1
    fi
    
    print_success "${joined_peers} peer(s) have joined channel ${channel_name}"
    
    # Check if peer containers are running
    if ! check_fabric_container "peer0.org1.ibn.vn"; then
        print_error "Peer containers are not running. Please start the network first."
        return 1
    fi
    
    # Step 0: Build chaincode (for Node.js/TypeScript chaincode)
    print_info "Step 0/7: Building chaincode..."
    
    # abs_chaincode_path already resolved above
    
    # Check if it's Node.js chaincode (has package.json)
    if [ -f "${abs_chaincode_path}/package.json" ]; then
        print_info "Detected Node.js chaincode, building..."
        
        # Check if npm is available
        if ! check_command npm; then
            print_warning "npm not found. Skipping build step. Make sure chaincode is already built."
        else
            # Build chaincode
            print_info "Installing dependencies..."
            (cd "${abs_chaincode_path}" && npm install)
            if [ $? -ne 0 ]; then
                print_error "Failed to install chaincode dependencies"
                return 1
            fi
            
            print_info "Building TypeScript to JavaScript..."
            (cd "${abs_chaincode_path}" && npm run build)
            if [ $? -ne 0 ]; then
                print_error "Failed to build chaincode"
                return 1
            fi
            
            # Check if dist/ directory exists
            if [ ! -d "${abs_chaincode_path}/dist" ]; then
                print_error "dist/ directory not found after build"
                return 1
            fi
            
            # Copy required files to dist/
            print_info "Copying required files to dist/..."
            if [ -f "${abs_chaincode_path}/msp-config.json" ]; then
                cp "${abs_chaincode_path}/msp-config.json" "${abs_chaincode_path}/dist/"
            else
                print_warning "msp-config.json not found, skipping..."
            fi
            
            if [ -f "${abs_chaincode_path}/package.json" ]; then
                cp "${abs_chaincode_path}/package.json" "${abs_chaincode_path}/dist/"
            else
                print_error "package.json not found"
                return 1
            fi
            
            # Regenerate package-lock.json in dist/ with --omit=dev
            print_info "Regenerating package-lock.json in dist/ (production only)..."
            (cd "${abs_chaincode_path}/dist" && rm -f package-lock.json && npm install --omit=dev --package-lock-only)
            if [ $? -ne 0 ]; then
                print_warning "Failed to regenerate package-lock.json, continuing anyway..."
            fi
            
            # Verify dist/ structure
            if [ ! -f "${abs_chaincode_path}/dist/index.js" ]; then
                print_error "index.js not found in dist/ after build"
                return 1
            fi
            
            if [ ! -f "${abs_chaincode_path}/dist/package.json" ]; then
                print_error "package.json not found in dist/"
                return 1
            fi
            
            print_success "Chaincode built successfully"
            # Update chaincode_path to point to dist/ directory
            chaincode_path="${abs_chaincode_path}/dist"
        fi
    else
        print_info "No package.json found, assuming chaincode is already built or using different language"
    fi
    
    echo ""
    
    # Step 1: Package chaincode
    print_info "Step 1/7: Packaging chaincode..."
    local package_file="${chaincode_name}_${chaincode_version}.tar.gz"
    local chaincode_label="${chaincode_name}_${chaincode_version}"
    local temp_package="/tmp/${package_file}"
    
    # Use first peer to package chaincode
    local first_peer="peer0.org1.ibn.vn"
    wait_for_fabric_container "${first_peer}"
    
    # Copy chaincode to peer container (use resolved path)
    print_info "Copying chaincode source to ${first_peer}..."
    local temp_chaincode_dir="/tmp/chaincode_${chaincode_name}"
    
    # Determine source path: if chaincode_path was updated to dist/, use it; otherwise use abs_chaincode_path
    local source_path
    if [[ "${chaincode_path}" == */dist ]]; then
        # Already pointing to dist/ (after build)
        source_path="${chaincode_path}"
    elif [[ "${chaincode_path}" == /* ]]; then
        # Absolute path (original)
        source_path="${chaincode_path}"
    else
        # Relative path, use abs_chaincode_path we calculated
        source_path="${abs_chaincode_path}"
    fi
    
    docker cp "${source_path}" "${first_peer}:${temp_chaincode_dir}" > /dev/null 2>&1
    
    if [ $? -ne 0 ]; then
        print_error "Failed to copy chaincode to peer container"
        return 1
    fi
    
    # Package chaincode inside peer container
    print_info "Creating chaincode package..."
    local package_output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
        -w /root \
        "${first_peer}" \
        peer lifecycle chaincode package "${package_file}" \
        --path "${temp_chaincode_dir}" \
        --lang node \
        --label "${chaincode_label}" 2>&1)
    
    if [ $? -ne 0 ]; then
        print_error "Failed to package chaincode: $package_output"
        docker exec "${first_peer}" rm -rf "${temp_chaincode_dir}" > /dev/null 2>&1
        return 1
    fi
    
    print_success "Chaincode packaged successfully: ${package_file}"
    
    # Copy package file to host
    docker cp "${first_peer}:/root/${package_file}" "${temp_package}" > /dev/null 2>&1
    
    # Cleanup temp chaincode dir
    docker exec "${first_peer}" rm -rf "${temp_chaincode_dir}" > /dev/null 2>&1
    
    echo ""
    
    # Step 2: Install chaincode on all peers
    print_info "Step 2/7: Installing chaincode on all peers..."
    local peers=("peer0.org1.ibn.vn:7051" "peer1.org1.ibn.vn:8051" "peer2.org1.ibn.vn:9051")
    local install_success=0
    local package_ids=()
    
    # Copy Admin MSP to all peers (required for lifecycle queries)
    print_info "Preparing Admin MSP for lifecycle operations..."
    local admin_msp_path="core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp"
    if [ ! -d "${admin_msp_path}" ]; then
        print_error "Admin MSP not found: ${admin_msp_path}"
        print_info "Please ensure crypto material is generated"
        return 1
    fi
    
    for peer_info in "${peers[@]}"; do
        IFS=':' read -r peer_name peer_port <<< "${peer_info}"
        if check_fabric_container "${peer_name}"; then
            docker cp "${admin_msp_path}" "${peer_name}:/tmp/admin_msp" > /dev/null 2>&1
            if [ $? -ne 0 ]; then
                print_warning "Failed to copy Admin MSP to ${peer_name}, continuing anyway..."
            fi
        fi
    done
    print_success "Admin MSP prepared for all peers"
    echo ""
    
    for peer_info in "${peers[@]}"; do
        IFS=':' read -r peer_name peer_port <<< "${peer_info}"
        
        print_info "Installing on ${peer_name}..."
        
        if ! check_fabric_container "${peer_name}"; then
            print_warning "Skipping ${peer_name} (container not running)"
            continue
        fi
        
        wait_for_fabric_container "${peer_name}"
        
        # Copy package to peer if not already there
        if [ -f "${temp_package}" ]; then
            docker cp "${temp_package}" "${peer_name}:/root/${package_file}" > /dev/null 2>&1
        fi
        
        # Install chaincode (can use peer MSP for install, but Admin MSP for query)
        local install_output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
            -e CORE_PEER_TLS_ENABLED=true \
            -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
            -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
            -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
            -w /root \
            "${peer_name}" \
            peer lifecycle chaincode install "${package_file}" 2>&1)
        
        if [ $? -eq 0 ]; then
            print_success "Installed on ${peer_name}"
            install_success=$((install_success + 1))
            
            # Try to extract package ID from install output first
            local pkg_id=""
            
            # Multiple formats to try:
            # 1. "Installed remotely: response:<status:200 payload:"\nEteaTraceCC_1.0:hash..."
            pkg_id=$(echo "$install_output" | grep -oE "${chaincode_label}:[a-f0-9]{64,}" | head -1)
            
            # 2. "Installed chaincode with package ID 'teaTraceCC_1.0:hash...'"
            if [ -z "$pkg_id" ]; then
                pkg_id=$(echo "$install_output" | grep -oE "package ID ['\"]?${chaincode_label}:[a-f0-9]+" | sed "s/.*package ID ['\"]\?//" | head -1)
            fi
            
            # 3. "Package ID: teaTraceCC_1.0:hash..."
            if [ -z "$pkg_id" ]; then
                pkg_id=$(echo "$install_output" | grep -i "Package ID" | grep -oE "${chaincode_label}:[a-f0-9]+" | head -1)
            fi
            
            # If not found in install output, try querying with Admin MSP
            if [ -z "$pkg_id" ]; then
                pkg_id=$(query_installed_chaincode "${peer_name}" "${chaincode_label}")
            fi
            
            if [ -n "$pkg_id" ]; then
                package_ids+=("${pkg_id}")
                print_info "Package ID on ${peer_name}: ${pkg_id}"
            else
                print_warning "Could not extract package ID from ${peer_name}, but install was successful"
            fi
        else
            print_error "Failed to install on ${peer_name}: $install_output"
        fi
    done
    
    if [ $install_success -eq 0 ]; then
        print_error "Failed to install chaincode on any peer"
        rm -f "${temp_package}"
        return 1
    fi
    
    if [ $install_success -lt ${#peers[@]} ]; then
        print_warning "Installed on ${install_success}/${#peers[@]} peers"
    else
        print_success "Installed on all ${#peers[@]} peers"
    fi
    
    # Get package ID (use first one)
    local package_id="${package_ids[0]}"
    if [ -z "$package_id" ]; then
        print_warning "Could not determine package ID, trying to query..."
        package_id=$(query_installed_chaincode "peer0.org1.ibn.vn" "${chaincode_label}")
    fi
    
    if [ -z "$package_id" ]; then
        print_error "Could not determine package ID. Please check installation logs."
        rm -f "${temp_package}"
        return 1
    fi
    
    print_success "Package ID: ${package_id}"
    echo ""
    
    # Step 3: Approve chaincode definition
    print_info "Step 3/7: Approving chaincode definition..."
    
    # Approve on first peer (representing organization) - Use Admin MSP
    local approve_output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
        -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
        -w /root \
        "peer0.org1.ibn.vn" \
        peer lifecycle chaincode approveformyorg \
        -o "orderer.ibn.vn:7050" \
        --channelID "${channel_name}" \
        --name "${chaincode_name}" \
        --version "${chaincode_version}" \
        --package-id "${package_id}" \
        --sequence "${sequence}" \
        --tls \
        --cafile "/etc/hyperledger/fabric/tls/ca.crt" \
        $( [ "$init_required" = "true" ] && echo "--init-required" || echo "" ) 2>&1)
    
    if [ $? -eq 0 ]; then
        print_success "Chaincode definition approved"
    else
        print_error "Failed to approve chaincode definition: $approve_output"
        rm -f "${temp_package}"
        return 1
    fi
    
    echo ""
    sleep 2
    
    # Step 4: Check commit readiness
    print_info "Step 4/7: Checking commit readiness..."
    if check_commit_readiness "peer0.org1.ibn.vn" "${channel_name}" "${chaincode_name}" "${chaincode_version}" "${sequence}"; then
        print_success "Chaincode is ready to commit"
    else
        print_warning "Commit readiness check failed, but continuing..."
    fi
    
    echo ""
    sleep 2
    
    # Step 5: Commit chaincode definition
    print_info "Step 5/7: Committing chaincode definition to channel..."
    
    # Use Admin MSP for commit (required for ACL)
    local commit_output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
        -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
        -w /root \
        "peer0.org1.ibn.vn" \
        peer lifecycle chaincode commit \
        -o "orderer.ibn.vn:7050" \
        --channelID "${channel_name}" \
        --name "${chaincode_name}" \
        --version "${chaincode_version}" \
        --sequence "${sequence}" \
        --tls \
        --cafile "/etc/hyperledger/fabric/tls/ca.crt" \
        $( [ "$init_required" = "true" ] && echo "--init-required" || echo "" ) 2>&1)
    
    if [ $? -eq 0 ]; then
        print_success "Chaincode committed to channel ${channel_name}"
    else
        print_error "Failed to commit chaincode: $commit_output"
        rm -f "${temp_package}"
        return 1
    fi
    
    echo ""
    sleep 2
    
    # Step 6: Verify deployment
    print_info "Step 6/7: Verifying chaincode deployment..."
    
    # Use Admin MSP for querycommitted (required for ACL)
    local query_output=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
        -e CORE_PEER_TLS_ENABLED=true \
        -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
        -e CORE_PEER_MSPCONFIGPATH="/tmp/admin_msp" \
        -e CORE_PEER_ADDRESS="peer0.org1.ibn.vn:7051" \
        -w /root \
        "peer0.org1.ibn.vn" \
        peer lifecycle chaincode querycommitted \
        --channelID "${channel_name}" \
        --name "${chaincode_name}" \
        --tls \
        --cafile "/etc/hyperledger/fabric/tls/ca.crt" 2>&1)
    
    if [ $? -eq 0 ]; then
        print_success "Chaincode deployment verified"
        echo ""
        print_info "Chaincode Information:"
        echo "$query_output" | grep -E "(Name|Version|Sequence|Endorsement Plugin|Validation Plugin)" || echo "$query_output"
    else
        print_warning "Could not verify deployment, but commit was successful"
    fi
    
    # Cleanup
    rm -f "${temp_package}"
    
    echo ""
    print_success "Chaincode ${chaincode_name} v${chaincode_version} deployed successfully to channel ${channel_name}!"
    return 0
}

# Bootstrap Fabric Network - Professional full setup
bootstrap_network() {
    print_header "Khởi Tạo Mạng Blockchain (Bootstrap)"
    echo ""
    
    local channel_name="ibnchannel"
    local chaincode_name="teaTraceCC"
    local chaincode_version="1.0"
    local chaincode_path="teaTraceCC"
    
    # Resolve chaincode path (fallback for different script locations)
    if [ ! -d "${chaincode_path}" ] && [ -d "../teaTraceCC" ]; then
        chaincode_path="../teaTraceCC"
    fi
    
    print_info "Full Network Bootstrap Configuration:"
    echo "  Channel: ${channel_name}"
    echo "  Chaincode: ${chaincode_name} v${chaincode_version}"
    echo "  Chaincode Path: ${chaincode_path}"
    echo ""
    
    # Validate prerequisites
    print_info "Validating prerequisites..."
    
    # Check if containers are running
    local required_containers=("orderer.ibn.vn" "peer0.org1.ibn.vn" "peer1.org1.ibn.vn" "peer2.org1.ibn.vn")
    local missing_containers=()
    
    for container in "${required_containers[@]}"; do
        if ! check_fabric_container "${container}"; then
            missing_containers+=("${container}")
        fi
    done
    
    if [ ${#missing_containers[@]} -gt 0 ]; then
        print_error "Required containers are not running:"
        for container in "${missing_containers[@]}"; do
            echo "  - ${container}"
        done
        print_info "Please start the network first using option 1 (Fresh Setup) or 2 (Normal Start)"
        return 1
    fi
    
    print_success "All required containers are running"
    echo ""
    
    # Step 1: Create channel
    print_header "Step 1/3: Create Channel"
    echo ""
    if ! create_fabric_channel "${channel_name}"; then
        print_error "Channel creation failed. Aborting bootstrap."
        return 1
    fi
    
    echo ""
    print_info "Waiting for channel to be fully created..."
    sleep 3
    
    # Step 2: Join peers
    print_header "Step 2/3: Join Peers to Channel"
    echo ""
    if ! join_fabric_peers "${channel_name}" 3; then
        print_error "Failed to join peers to channel"
        print_error "Cannot deploy chaincode without peers in channel"
        print_info "Please fix peer joining issues first, then deploy chaincode manually (option 5)"
        return 1
    fi
    
    # Verify at least one peer has joined
    print_info "Verifying peers have joined channel..."
    local joined_count=0
    local peers=("peer0.org1.ibn.vn:7051" "peer1.org1.ibn.vn:8051" "peer2.org1.ibn.vn:9051")
    
    for peer_info in "${peers[@]}"; do
        IFS=':' read -r peer_name peer_port <<< "${peer_info}"
        if check_fabric_container "${peer_name}"; then
            local channel_list=$(docker exec -e CORE_PEER_LOCALMSPID="Org1MSP" \
                -e CORE_PEER_TLS_ENABLED=true \
                -e CORE_PEER_TLS_ROOTCERT_FILE="/etc/hyperledger/fabric/tls/ca.crt" \
                -e CORE_PEER_MSPCONFIGPATH="/etc/hyperledger/fabric/msp" \
                -e CORE_PEER_ADDRESS="${peer_name}:${peer_port}" \
                -w /root \
                "${peer_name}" \
                peer channel list 2>&1 | grep -i "${channel_name}" || echo "")
            
            if [ -n "$channel_list" ]; then
                joined_count=$((joined_count + 1))
                print_success "${peer_name} is in channel ${channel_name}"
            fi
        fi
    done
    
    if [ $joined_count -eq 0 ]; then
        print_error "No peers have joined channel ${channel_name}"
        print_error "Cannot deploy chaincode without peers in channel"
        print_info "Please join peers first (option 4), then deploy chaincode (option 5)"
        return 1
    fi
    
    print_success "${joined_count} peer(s) have joined channel ${channel_name}"
    echo ""
    print_info "Waiting for peers to sync..."
    sleep 3
    
    # Step 3: Deploy chaincode
    print_header "Step 3/3: Deploy Chaincode"
    echo ""
    if ! deploy_fabric_chaincode "${channel_name}" "${chaincode_name}" "${chaincode_version}" "${chaincode_path}" "1" "false"; then
        print_error "Chaincode deployment failed"
        print_warning "Network is partially set up. Channel created and peers joined, but chaincode not deployed."
        return 1
    fi
    
    echo ""
    print_header "Bootstrap Summary"
    echo ""
    print_success "Network bootstrap completed successfully!"
    echo ""
    print_info "Network Status:"
    echo "  ✓ Channel '${channel_name}' created"
    echo "  ✓ Peers joined to channel"
    echo "  ✓ Chaincode '${chaincode_name}' v${chaincode_version} deployed"
    echo ""
    print_info "Next steps:"
    echo "  • Test chaincode using API Gateway endpoints"
    echo "  • Monitor network using Grafana dashboard"
    echo "  • View logs using option 7 (View Logs)"
    echo ""
    
    return 0
}

# Cleanup environment
cleanup_environment() {
    print_header "Dọn Dẹp Môi Trường (Fresh Setup)"
    echo ""
    
    print_info "Đang dừng các containers..."
    docker compose down -v  # -v to remove named volumes
    
    print_info "Đang xóa các volumes cụ thể của Fabric..."
    # Remove specific named volumes to ensure clean state
    docker volume rm -f orderer.ibn.vn orderer1.ibn.vn orderer2.ibn.vn 2>/dev/null || true
    docker volume rm -f peer0.org1.ibn.vn peer1.org1.ibn.vn peer2.org1.ibn.vn 2>/dev/null || true
    docker volume rm -f couchdb0 couchdb1 couchdb2 2>/dev/null || true
    
    print_info "Đang xóa dangling volumes..."
    docker volume prune -f
    
    print_info "Đang xóa crypto material cũ..."
    rm -rf core/organizations
    rm -rf core/system-genesis-block
    rm -rf core/channel-artifacts
    
    print_info "Đang xóa chaincode build artifacts..."
    rm -rf teaTraceCC/dist
    rm -rf teaTraceCC/node_modules
    
    print_success "✅ Đã dọn dẹp xong! Môi trường hoàn toàn sạch cho cài đặt mới."
    echo ""
    print_warning "⚠️  Lưu ý: Tất cả dữ liệu blockchain cũ đã bị xóa!"
    echo ""
}

print_menu_option() {
    local index="$1"
    local icon="$2"
    local title="$3"
    local desc="$4"
    local subtext="$5"
    
    # Calculate visual length of the main line
    # [index] icon title (desc)
    local main_text="  [${index}] ${icon} ${title} ${desc}"
    local text_len=${#main_text}
    
    # Correction for emoji (assume +1 visual width per emoji)
    # Most emojis are 1 char in string length but 2 cells wide
    local visual_len=$((text_len + 1))
    
    # Target inner width is 60 chars
    # We print: ║${main_text}${padding}║
    # So padding = 60 - visual_len
    local padding=$((60 - visual_len))
    
    if [ $padding -lt 0 ]; then padding=0; fi
    local spaces=$(printf "%${padding}s")
    
    echo -e "${CYAN}${BOLD}║${NC}${GREEN}${BOLD}  [${index}] ${icon} ${title}${NC} ${desc}${spaces}${CYAN}${BOLD}║${NC}"
    
    # Print subtext if exists
    if [ -n "$subtext" ]; then
        local sub_text="      ${YELLOW}→ ${subtext}${NC}"
        # Strip colors for length
        local stripped_sub=$(echo -e "$sub_text" | sed 's/\x1b\[[0-9;]*m//g')
        local sub_len=${#stripped_sub}
        local sub_padding=$((60 - sub_len))
        if [ $sub_padding -lt 0 ]; then sub_padding=0; fi
        local sub_spaces=$(printf "%${sub_padding}s")
        
        echo -e "${CYAN}${BOLD}║${NC}${sub_text}${sub_spaces}${CYAN}${BOLD}║${NC}"
        echo -e "${CYAN}${BOLD}║${NC}                                                            ${CYAN}${BOLD}║${NC}"
    fi
}

show_main_menu() {
    print_header "MENU CHÍNH"
    echo ""
    echo -e "${CYAN}${BOLD}╔${DIVIDER}╗${NC}"
    
    print_menu_option "1" "🚀" "Fresh Setup" "(Cài đặt mới toàn bộ)" "Xóa sạch, tạo crypto, start containers, setup network"
    print_menu_option "2" "▶️" "Normal Start" "(Khởi động lại)" "Giữ nguyên dữ liệu, chỉ start containers"
    print_menu_option "3" "🌐" "Create Channel" "(Chỉ tạo channel)"
    print_menu_option "4" "🔗" "Join Peers" "(Chỉ join peers)"
    print_menu_option "5" "📜" "Deploy Chaincode" "(Chỉ deploy chaincode)"
    print_menu_option "6" "🛑" "Stop Project" "(Dừng toàn bộ)"
    print_menu_option "7" "👀" "View Error Logs" "(Xem logs lỗi)" "Chỉ hiển thị ERROR, WARN, PANIC"
    print_menu_option "8" "👑" "Create Default Admin" "(Tạo admin để đăng nhập hệ thống)"
    print_menu_option "9" "❌" "Exit" "(Thoát)"
    
    echo -e "${CYAN}${BOLD}╚${DIVIDER}╝${NC}"
    echo ""
    echo -ne "${BOLD}➜ Nhập lựa chọn của bạn (1-9): ${NC}"
}

# Create default admin user (DEV/DEMO only)
# - Tạo admin@ibn.vn trong database (nếu chưa có hoặc cập nhật lại)
# - Kiểm tra Admin MSP cho Org1 (blockchain identity)
# - Thử login qua Backend API để xác nhận tài khoản dùng được
create_default_admin() {
    print_header "Tạo Admin Mặc Định (DEV/DEMO)"
    echo ""

    # 1. Kiểm tra Postgres bắt buộc phải RUNNING
    print_info "Đang kiểm tra containers bắt buộc..."
    
    local required_containers=("ibn-postgres")
    local missing=()

    for c in "${required_containers[@]}"; do
        local status
        status=$(docker inspect -f '{{.State.Status}}' "$c" 2>/dev/null || echo "not-found")
        if [ "$status" != "running" ]; then
            missing+=("$c (status: $status)")
        fi
    done

    if [ ${#missing[@]} -gt 0 ]; then
        print_error "Các containers sau chưa chạy hoặc không tồn tại:"
        for c in "${missing[@]}"; do
            echo "  - $c"
        done
        echo ""
        print_info "Hãy chọn option 1 (Fresh Setup) hoặc 2 (Normal Start) để khởi động services trước."
        return 1
    fi
    
    print_success "✓ Postgres đang chạy"
    
    # 1.5. Backend là TUỲ CHỌN (chỉ để test login)
    local backend_status
    backend_status=$(docker inspect -f '{{.State.Status}}' "ibn-backend" 2>/dev/null || echo "not-found")
    if [ "$backend_status" != "running" ]; then
        print_warning "Backend (ibn-backend) chưa chạy (status: $backend_status)."
        print_info "Vẫn tiếp tục tạo admin trong database, nhưng KHÔNG test đăng nhập qua API được."
    else
        print_info "Đang đợi backend sẵn sàng (max 30s)..."
        local wait_time=0
        while [ $wait_time -lt 30 ]; do
            if curl -s http://localhost:9900/health > /dev/null 2>&1; then
                print_success "✓ Backend sẵn sàng"
                break
            fi
            echo -n "."
            sleep 2
            wait_time=$((wait_time + 2))
        done
        echo ""
        
        if [ $wait_time -ge 30 ]; then
            print_warning "Backend chưa sẵn sàng sau 30s, nhưng sẽ tiếp tục thử tạo admin..."
        fi
    fi

    # 2. Kiểm tra Admin MSP trên filesystem (blockchain identity)
    local admin_msp_path="core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp"
    if [ -d "$admin_msp_path" ] && [ -n "$(ls -A "$admin_msp_path" 2>/dev/null)" ]; then
        print_success "Admin MSP (Org1) đã tồn tại: $admin_msp_path"
    else
        print_warning "Không tìm thấy Admin MSP tại: $admin_msp_path"
        print_info "Admin MSP sẽ được tạo lại nếu bạn chạy Fresh Setup (option 1)."
    fi

    echo ""
    
    # 3. Kiểm tra và tạo schemas + tables nếu chưa có
    print_info "Đang kiểm tra database schema..."
    
    # Đảm bảo extension pgcrypto tồn tại (cần cho gen_random_uuid)
    print_info "Đảm bảo extension 'pgcrypto' tồn tại trong database..."
    if ! docker exec ibn-postgres psql -U postgres -d ibn_gateway -c "CREATE EXTENSION IF NOT EXISTS pgcrypto;" >/dev/null 2>&1; then
        print_warning "Không thể tạo extension pgcrypto (có thể đã tồn tại hoặc thiếu quyền)."
    else
        print_success "✓ Extension 'pgcrypto' đã được đảm bảo tồn tại."
    fi
    
    # Luôn kiểm tra schema + table, nếu thiếu thì apply migrations
    # 3.1. Kiểm tra schema auth
    local schema_exists
    schema_exists=$(docker exec ibn-postgres psql -U gateway -d ibn_gateway -t -c "SELECT EXISTS(SELECT 1 FROM information_schema.schemata WHERE schema_name = 'auth');" | tr -d ' ')

    # 3.2. Kiểm tra table auth.users
    local table_exists
    table_exists=$(docker exec ibn-postgres psql -U gateway -d ibn_gateway -t -c "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'auth' AND table_name = 'users');" | tr -d ' ')

    if [ "$schema_exists" != "t" ] || [ "$table_exists" != "t" ]; then
        if [ "$schema_exists" != "t" ]; then
            print_warning "Schema 'auth' chưa tồn tại. Đang chạy toàn bộ migrations..."
        else
            print_warning "Schema 'auth' đã tồn tại nhưng table 'auth.users' thiếu. Đang chạy lại migrations..."
        fi

        # Apply migrations thủ công (backend không có auto-migration)
        local migration_count=0
        local failed_migrations=()
        
        for migration in backend/migrations/*.up.sql; do
            if [ -f "$migration" ]; then
                local migration_name
                migration_name=$(basename "$migration")
                print_info "  ➜ Applying $migration_name..."
                
                if docker exec -i ibn-postgres psql -U gateway -d ibn_gateway < "$migration" > "/tmp/migration_${migration_count}.log" 2>&1; then
                    print_success "    ✓ $migration_name"
                    ((migration_count++))
                else
                    print_error "    ✗ $migration_name failed"
                    failed_migrations+=("$migration_name")
                fi
            fi
        done
        
        echo ""
        if [ ${#failed_migrations[@]} -eq 0 ]; then
            print_success "✓ Đã apply $migration_count migrations thành công"
        else
            print_error "Có ${#failed_migrations[@]} migrations thất bại:"
            for failed in "${failed_migrations[@]}"; do
                echo "  - $failed"
            done
            print_info "Xem chi tiết lỗi migrations trong container postgres (ibn-postgres):"
            echo "  docker exec -it ibn-postgres bash"
            echo "  ls -la /tmp/migration_*.log"
            return 1
        fi

        # Sau khi migrate lại, re-check table users
        table_exists=$(docker exec ibn-postgres psql -U gateway -d ibn_gateway -t -c "SELECT EXISTS(SELECT 1 FROM information_schema.tables WHERE table_schema = 'auth' AND table_name = 'users');" | tr -d ' ')
    else
        print_success "✓ Schema 'auth' và các bảng cơ bản đã tồn tại"
    fi
    
    if [ "$table_exists" != "t" ]; then
        print_error "Table 'auth.users' vẫn không tồn tại sau khi chạy migrations!"
        print_info "Vui lòng kiểm tra logs migration trong container postgres (ibn-postgres):"
        echo "  docker exec -it ibn-postgres bash"
        echo "  ls -la /tmp/migration_*.log"
        return 1
    fi
    
    print_success "✓ Table 'auth.users' tồn tại"
    
    # 4. Sinh bcrypt hash cho password admin123 (compatible với Go bcrypt)
    print_info "Đang sinh password hash..."
    local admin_email="admin@ibn.vn"
    local admin_username="admin"
    local admin_password_plain="admin123"
    
    # Sinh hash bằng Python bcrypt với prefix $2a$ (Go bcrypt compatible)
    local password_hash
    if check_command python3; then
        password_hash=$(python3 -c "import bcrypt; print(bcrypt.hashpw(b'$admin_password_plain', bcrypt.gensalt(rounds=10, prefix=b'2a')).decode())" 2>/dev/null)
    fi
    
    # Fallback: Dùng hash có sẵn (đã test với Go bcrypt)
    if [ -z "$password_hash" ]; then
        print_warning "Không thể sinh hash bằng Python, dùng hash có sẵn"
        password_hash='$2a$10$9ztl4PCS90ONiUmXSL7e0eIflXycCj5vrvutzVRUtyU.t8oJwB09C'
    fi
    
    print_success "✓ Password hash đã sẵn sàng"

    # 5. Tạo admin trong PostgreSQL
    print_info "Đang tạo/cập nhật admin user..."

    # Insert vào schema auth.users (không phải public.users)
    # Phải escape $ trong password hash để tránh shell interpretation
    docker exec -i ibn-postgres psql -U gateway -d ibn_gateway <<EOSQL
INSERT INTO auth.users (email, username, password_hash, full_name, role, msp_id, is_active, email_verified)
VALUES ('$admin_email', '$admin_username', '$password_hash', 'Admin User', 'admin', 'Org1MSP', TRUE, TRUE)
ON CONFLICT (email) DO UPDATE SET
    username = EXCLUDED.username,
    password_hash = EXCLUDED.password_hash,
    full_name = EXCLUDED.full_name,
    role = EXCLUDED.role,
    msp_id = EXCLUDED.msp_id,
    is_active = EXCLUDED.is_active,
    email_verified = EXCLUDED.email_verified,
    updated_at = CURRENT_TIMESTAMP;
EOSQL

    if [ $? -ne 0 ]; then
        print_error "Không thể tạo/cập nhật admin trong database."
        return 1
    fi

    print_success "✓ Đã tạo/cập nhật admin trong database"

    # 5. Thử login qua Backend API để xác nhận (nếu backend đang chạy)
    echo ""
    if [ "$backend_status" != "running" ]; then
        print_warning "Bỏ qua bước test đăng nhập vì backend (ibn-backend) chưa chạy."
    else
        print_info "Đang kiểm tra đăng nhập qua Backend API..."

        if ! check_command curl; then
            print_warning "curl không tồn tại, bỏ qua bước kiểm tra đăng nhập."
        else
            local login_response
            local login_status
            
            login_response=$(curl -s -w "\n%{http_code}" \
                "http://localhost:9900/api/v1/auth/login" \
                -H "Content-Type: application/json" \
                -d "{\"email\":\"$admin_email\",\"password\":\"$admin_password_plain\"}" 2>/dev/null || echo -e "\n000")
            
            login_status=$(echo "$login_response" | tail -1)
            local response_body
            response_body=$(echo "$login_response" | head -n -1)

            if [ "$login_status" = "200" ]; then
                print_success "✓ Đăng nhập thành công!"
                
                # Extract access token để verify
                local access_token
                access_token=$(echo "$response_body" | grep -o '"access_token":"[^"]*"' | cut -d'"' -f4)
                if [ -n "$access_token" ]; then
                    print_success "✓ Nhận được access token (${#access_token} chars)"
                fi
            else
                print_error "✗ Đăng nhập thất bại (HTTP $login_status)"
                echo ""
                print_info "Response body:"
                echo "$response_body" | head -5
                echo ""
                print_info "Có thể backend chưa hoàn toàn sẵn sàng. Vui lòng thử lại sau hoặc kiểm tra logs:"
                echo "  docker logs ibn-backend"
            fi
        fi
    fi

    echo ""
    print_header "Thông Tin Tài Khoản Admin (DEV/DEMO)"
    echo ""
    echo -e "  ${GREEN}📧 Email:${NC}       $admin_email"
    echo -e "  ${GREEN}👤 Username:${NC}    $admin_username"
    echo -e "  ${GREEN}🔑 Password:${NC}    $admin_password_plain"
    echo ""
    echo -e "  ${GREEN}🌐 API Endpoint:${NC}"
    echo "     http://localhost:9900/api/v1/auth/login"
    echo ""
    print_warning "⚠ Tài khoản này CHỈ dùng cho môi trường DEV/DEMO. Không dùng cho production."
    echo ""
    
    # Thêm hướng dẫn test
    print_info "Để test đăng nhập, copy và paste lệnh sau:"
    echo ""
    echo -e "${CYAN}curl -X POST http://localhost:9900/api/v1/auth/login \\${NC}"
    echo -e "${CYAN}  -H 'Content-Type: application/json' \\${NC}"
    echo -e "${CYAN}  -d '{\"email\":\"$admin_email\",\"password\":\"$admin_password_plain\"}'${NC}"
    echo ""
}

perform_fresh_setup() {
    print_info "Đang thực hiện Fresh Setup..."
    cleanup_environment
    
    print_info "Đang tạo lại crypto material và genesis block..."
    generate_crypto_material "core"
    generate_genesis_block "core"
    
    perform_normal_start "fresh"
}

perform_normal_start() {
    local mode=$1
    
    # Install dependencies if in Development mode
    if [ "$USAGE_MODE" == "development" ]; then
        install_project_dependencies
    fi

    print_info "Đang khởi động dự án..."
    echo ""
    
    # Check Docker access again before running
    if ! docker ps &> /dev/null; then
        print_error "Không thể truy cập Docker daemon"
        return 1
    fi
    
    # Run docker compose up -d (build and start)
    print_info "Đang build và khởi động các services..."
    echo ""
    if docker compose up -d --build 2>&1 | grep -E "(Built|Created|Started|Healthy|Error|Pulling|Downloaded)" | tail -20; then
        echo ""
        print_success "Dự án đã được build và khởi động thành công!"
        echo ""
        print_info "Đang chờ services khởi động (30 giây)..."
        sleep 30
        echo ""
        
        # Wait for Fabric network to be ready (TLS handshake, Raft cluster, gossip)
        if [ "$mode" == "fresh" ]; then
            print_header "Verifying Fabric Network Health"
            echo ""
            wait_for_fabric_network
            echo ""
        fi
        
        # Check services status
        print_header "Trạng Thái Services"
        echo ""
        docker compose ps
        echo ""
        
        # Bootstrap Network if fresh setup or explicitly requested (implied in fresh setup flow)
        if [ "$mode" == "fresh" ]; then
             bootstrap_network
        fi
        
        echo ""
        
        # Check if services are healthy
        print_info "Đang kiểm tra health của các services..."
        echo ""
        
        # Wait a bit more for services to be ready
        sleep 10
        
        print_header "Hướng Dẫn Sử Dụng"
        echo ""
        print_info "Truy cập services:"
        echo "  • Frontend: http://localhost:9999"
        echo "  • Backend API: http://localhost:9900"
        echo "  • API Gateway: http://localhost:9805"
        echo "  • Grafana: http://localhost:9300"
        echo "  • Prometheus: http://localhost:9901"
        echo ""
        print_success "🎉 Dự án đã sẵn sàng sử dụng!"
        echo ""
    else
        print_error "Không thể khởi động dự án. Vui lòng kiểm tra logs."
    fi
}

# Main installation script
main() {
    print_banner
    
    # Detect OS automatically
    DETECTED_OS=$(get_os)
    
    # Ask user to confirm or select OS
    if [[ "$DETECTED_OS" == "unknown" ]] || [[ "$DETECTED_OS" == "windows" ]]; then
        echo ""
        print_warning "Không thể tự động phát hiện hệ điều hành, hoặc phát hiện Windows."
        OS=$(ask_os)
    else
        echo ""
        echo -e "${GREEN}✓${NC} ${BLUE}Đã phát hiện hệ điều hành: ${YELLOW}$DETECTED_OS${NC}"
        echo ""
        echo -e "${YELLOW}┌──────────────────────────────────────────────────────┐${NC}"
        echo -e "${YELLOW}│${NC}  ${BLUE}Bạn có muốn tiếp tục với hệ điều hành này? (Y/n)${NC}  ${YELLOW}  │${NC}"
        echo -e "${YELLOW}└──────────────────────────────────────────────────────┘${NC}"
        echo -ne "${GREEN}➜${NC} "
        read -n 1 -r
        echo
        if [[ $REPLY =~ ^[Nn]$ ]]; then
            OS=$(ask_os)
        else
            OS=$DETECTED_OS
        fi
    fi
    
    echo ""
    
    # Handle Windows separately
    if [[ "$OS" == "windows" ]]; then
        handle_windows
        exit 0
    fi
    
    # Check Docker first - this is the main requirement
    print_header "Kiểm Tra Docker"
    echo ""
    
    local docker_status=$(check_tool_status "docker" "" "$DOCKER_VERSION_MIN")
    local docker_installed=false
    
    if [[ "$docker_status" == ok:* ]]; then
        local docker_version=${docker_status#ok:}
        print_success "Docker đã được cài đặt: $docker_version"
        
        # Check Docker permissions
        if ! docker ps &> /dev/null; then
            print_warning "Không thể truy cập Docker daemon (permission denied)"
            echo ""
            print_info "Đang kiểm tra quyền truy cập Docker..."
            
            # Check if user is in docker group
            if groups | grep -q docker; then
                print_warning "User đã trong docker group nhưng vẫn không truy cập được"
                print_info "  → Có thể cần logout và login lại, hoặc restart Docker service"
            else
                print_warning "User chưa được thêm vào docker group"
                echo ""
                echo -e "${YELLOW}┌──────────────────────────────────────────────────────┐${NC}"
                echo -e "${YELLOW}│${NC}  ${BLUE}Bạn có muốn thêm user vào docker group? (Y/n)${NC}        ${YELLOW}│${NC}"
                echo -e "${YELLOW}└──────────────────────────────────────────────────────┘${NC}"
                echo -ne "${GREEN}➜${NC} "
                read -n 1 -r
                echo
                if [[ ! $REPLY =~ ^[Nn]$ ]]; then
                    if [[ "$OS" == "linux" ]] || [[ "$OS" == "wsl" ]]; then
                        sudo usermod -aG docker $USER
                        print_success "Đã thêm user vào docker group"
                        print_warning "⚠ QUAN TRỌNG: Bạn cần áp dụng thay đổi để sử dụng Docker"
                        echo ""
                        print_info "Có 2 cách để áp dụng:"
                        echo "  1. ${GREEN}Logout và login lại${NC} (khuyến nghị)"
                        echo "  2. ${GREEN}Chạy: newgrp docker${NC} (tạm thời trong session hiện tại)"
                        echo ""
                        echo -e "${YELLOW}┌──────────────────────────────────────────────────────┐${NC}"
                        echo -e "${YELLOW}│${NC}  ${BLUE}Bạn muốn làm gì?${NC}                                    ${YELLOW}│${NC}"
                        echo -e "${YELLOW}│${NC}  ${GREEN}(1)${NC} Chạy 'newgrp docker' và tiếp tục script          ${YELLOW}│${NC}"
                        echo -e "${YELLOW}│${NC}  ${GREEN}(2)${NC} Thoát và logout/login lại, rồi chạy lại script   ${YELLOW}│${NC}"
                        echo -e "${YELLOW}└──────────────────────────────────────────────────────┘${NC}"
                        echo -ne "${GREEN}➜${NC} "
                        read -n 1 -r
                        echo
                        if [[ $REPLY == "1" ]]; then
                            print_info "Đang chạy newgrp docker..."
                            echo ""
                            print_warning "Lưu ý: newgrp sẽ tạo shell mới. Script sẽ tiếp tục trong shell đó."
                            echo ""
                            
                            # Find project root and script path before newgrp
                            local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
                            local script_path="$script_dir/setup.sh"
                            local current_project_root=""
                            
                            if [ -f "docker-compose.yml" ]; then
                                current_project_root="$(pwd)"
                            elif [ -f "$script_dir/../docker-compose.yml" ]; then
                                current_project_root="$(cd "$script_dir/.." && pwd)"
                            elif [ -f "../docker-compose.yml" ]; then
                                current_project_root="$(cd .. && pwd)"
                            else
                                current_project_root="$(pwd)"
                            fi
                            
                            # Use sg (substitute group) instead of newgrp for better control
                            # sg can run a command directly with docker group privileges
                            print_info "Đang chạy script với quyền docker group..."
                            echo ""
                            
                            # Create command string to run
                            local setup_cmd="cd '$current_project_root' && export AUTO_SETUP=true && bash '$script_path'"
                            
                            # Use sg docker to run the script with docker group privileges
                            sg docker -c "$setup_cmd"
                            
                            exit 0
                        else
                            print_info "Vui lòng logout và login lại, sau đó chạy lại script:"
                            echo "  ${GREEN}bash scripts/setup.sh${NC}"
                            echo ""
                            print_info "Hoặc chạy: ${GREEN}newgrp docker${NC} và chạy lại script"
                            echo ""
                            exit 0
                        fi
                    fi
                fi
            fi
            
            echo ""
            print_error "Không thể tiếp tục vì không có quyền truy cập Docker"
            print_info "Vui lòng:"
            echo "  1. Logout và login lại (nếu đã thêm vào docker group)"
            echo "  2. Hoặc chạy: ${GREEN}newgrp docker${NC}"
            echo "  3. Hoặc chạy script với sudo (không khuyến nghị)"
            echo ""
            exit 1
        else
            print_success "Quyền truy cập Docker: OK"
        fi
        
        if check_command docker-compose || docker compose version &> /dev/null; then
            print_success "Docker Compose đã được cài đặt"
            docker_installed=true
        else
            print_warning "Docker Compose chưa được cài đặt. Đang cài đặt..."
            install_docker_compose
            docker_installed=true
        fi
    elif [[ "$docker_status" == outdated:* ]]; then
        local docker_version=${docker_status#outdated:}
        print_warning "Docker version $docker_version có thể cũ. Khuyến nghị: $DOCKER_VERSION_MIN"
        echo ""
        echo -e "${YELLOW}┌──────────────────────────────────────────────────────┐${NC}"
        echo -e "${YELLOW}│${NC}  ${BLUE}Bạn có muốn cập nhật Docker? (y/N)${NC}                    ${YELLOW}│${NC}"
        echo -e "${YELLOW}└──────────────────────────────────────────────────────┘${NC}"
        echo -ne "${GREEN}➜${NC} "
        read -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            install_docker
            docker_installed=true
        else
            print_info "Giữ nguyên phiên bản Docker hiện tại"
            docker_installed=true
        fi
    else
        print_warning "Docker chưa được cài đặt"
        echo ""
        echo -e "${YELLOW}┌──────────────────────────────────────────────────────┐${NC}"
        echo -e "${YELLOW}│${NC}  ${BLUE}Bạn có muốn cài đặt Docker ngay bây giờ? (Y/n)${NC}        ${YELLOW}  │${NC}"
        echo -e "${YELLOW}└──────────────────────────────────────────────────────┘${NC}"
        echo -ne "${GREEN}➜${NC} "
        read -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            install_docker
            docker_installed=true
        else
            print_error "Docker là bắt buộc để chạy dự án. Vui lòng cài đặt Docker trước."
            exit 1
        fi
    fi
    
    echo ""
    
    # Find project root directory (where docker-compose.yml is located)
    local script_dir="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"
    local project_root=""
    
    # Try to find docker-compose.yml from current directory
    if [ -f "docker-compose.yml" ]; then
        project_root="$(pwd)"
    # Try to find from script directory (scripts/setup.sh -> root)
    elif [ -f "$script_dir/../docker-compose.yml" ]; then
        project_root="$(cd "$script_dir/.." && pwd)"
    # Try to find from parent directory
    elif [ -f "../docker-compose.yml" ]; then
        project_root="$(cd .. && pwd)"
    else
        # Search in parent directories (up to 3 levels)
        local search_dir="$(pwd)"
        for i in {1..3}; do
            if [ -f "$search_dir/docker-compose.yml" ]; then
                project_root="$search_dir"
                break
            fi
            search_dir="$(dirname "$search_dir")"
        done
    fi
    
    # If still not found, use current directory and warn
    if [ -z "$project_root" ]; then
        project_root="$(pwd)"
        print_warning "Không tìm thấy docker-compose.yml trong thư mục hiện tại hoặc thư mục cha."
        print_info "Đang sử dụng thư mục hiện tại: $project_root"
        print_info "Vui lòng chạy script từ thư mục root của dự án (nơi có docker-compose.yml)"
        echo ""
    else
        # Change to project root
        cd "$project_root" || exit 1
        print_info "Đã tìm thấy thư mục root của dự án: $project_root"
        echo ""
    fi
    
    # Check project readiness
    print_header "Kiểm Tra Dự Án"
    echo ""
    
    local project_ready=true
    local missing_items=()
    
    # Check core directory (Fabric certificates)
    if [ ! -d "core/organizations" ] || [ -z "$(ls -A core/organizations)" ]; then
        print_warning "Thư mục core/organizations không tồn tại hoặc trống"
        print_info "  → Lần đầu chạy dự án, sẽ tự động tạo Fabric certificates."
        print_info "  → Nếu bạn đã từng chạy network trước đó, nên chọn Option 1 (Fresh Setup)."
        echo ""

        if generate_crypto_material "core"; then
            project_ready=true # Reset to true if generation succeeded, will be checked again
        else
            project_ready=false
            missing_items+=("core/organizations (Fabric certificates generation failed)")
        fi
    else
        print_success "Thư mục core/organizations: OK"
        
        # Check critical certificate paths
        local cert_checks=(
            "core/organizations/peerOrganizations/org1.ibn.vn/users/Admin@org1.ibn.vn/msp"
            "core/organizations/peerOrganizations/org1.ibn.vn/peers/peer0.org1.ibn.vn/msp"
            "core/organizations/ordererOrganizations/ibn.vn/orderers/orderer.ibn.vn/msp"
        )
        
        local cert_missing=false
        for cert_path in "${cert_checks[@]}"; do
            if [ ! -d "$cert_path" ] || [ -z "$(ls -A "$cert_path" 2>/dev/null)" ]; then
                print_warning "  Thiếu hoặc trống: $cert_path"
                cert_missing=true
            fi
        done
        
        if [ "$cert_missing" = true ]; then
            print_error "Phát hiện thiếu sót trong cấu trúc certificates hiện tại."
            print_warning "⚠ Vì lý do an toàn TLS, script sẽ KHÔNG tự động regenerate certificates ở chế độ này."
            print_info "  → Vui lòng chọn Option 1 (Fresh Setup) trong menu chính để:"
            print_info "     - Dừng toàn bộ containers"
            print_info "     - Xoá volumes cũ (ledger, state, Raft metadata)"
            print_info "     - Tạo lại crypto + genesis block đồng bộ"
            echo ""
            project_ready=false
            missing_items+=("Fabric certificates (inconsistent structure - run Fresh Setup)")}
        fi
    fi
    
    # Check genesis block (required for orderers)
    if [ ! -f "core/system-genesis-block/genesis.block" ]; then
        print_warning "File core/system-genesis-block/genesis.block không tồn tại"
        print_warning "⚠ Script sẽ không tự tạo genesis block ở bước kiểm tra này để tránh lệch TLS."
        print_info "  → Vui lòng chọn Option 1 (Fresh Setup) để generate genesis block cùng crypto material."
        project_ready=false
        missing_items+=("core/system-genesis-block/genesis.block (Missing - run Fresh Setup)")
    else
        print_success "Genesis block: OK"
    fi
    
    # Check Fabric config directory
    if [ ! -d "core/config" ]; then
        print_warning "Thư mục core/config không tồn tại"
        print_info "  → Cần có config files cho peers và orderers"
        project_ready=false
        missing_items+=("core/config")
    else
        if [ ! -f "core/config/core.yaml" ] && [ ! -f "core/config/orderer.yaml" ]; then
            print_warning "  Thiếu config files trong core/config"
            project_ready=false
            missing_items+=("core/config/*.yaml files")
        else
            print_success "Fabric config: OK"
        fi
    fi
    
    # Check Fabric CA config (optional but recommended)
    if [ ! -d "core/fabric-ca/org1" ]; then
        print_warning "Thư mục core/fabric-ca/org1 không tồn tại"
        print_info "  → Fabric CA service có thể không hoạt động đúng"
        # Don't mark as not ready, CA is optional if certificates are already generated
    else
        print_success "Fabric CA config: OK"
    fi
    
    # Check migrations
    if [ ! -d "backend/migrations" ]; then
        print_warning "Thư mục backend/migrations không tồn tại"
        project_ready=false
        missing_items+=("backend/migrations")
    else
        local migration_count=$(find backend/migrations -name "*.up.sql" 2>/dev/null | wc -l)
        if [ "$migration_count" -gt 0 ]; then
            print_success "Database migrations: $migration_count files found"
        else
            print_warning "Không tìm thấy migration files"
            project_ready=false
            missing_items+=("Migration files")
        fi
    fi
    
    # Check docker-compose.yml
    if [ ! -f "docker-compose.yml" ]; then
        print_error "File docker-compose.yml không tồn tại"
        project_ready=false
        missing_items+=("docker-compose.yml")
    else
        print_success "docker-compose.yml: OK"
    fi
    
    echo ""
    
    # Main Menu Loop
    while true; do
        show_main_menu
        read -r choice
        echo ""
        
        case $choice in
            1)
                perform_fresh_setup
                ;;
            2)
                print_info "Đang thực hiện Normal Start..."
                perform_normal_start
                ;;
            3)
                print_info "Đang tạo channel 'ibnchannel'..."
                create_fabric_channel "ibnchannel"
                ;;
            4)
                print_info "Đang join peers vào channel 'ibnchannel'..."
                join_fabric_peers "ibnchannel"
                ;;
            5)
                print_info "Đang deploy chaincode..."
                deploy_fabric_chaincode "ibnchannel" "teaTraceCC" "1.0" "teaTraceCC"
                ;;
            6)
                print_info "Đang dừng toàn bộ dự án..."
                docker compose down
                print_success "Đã dừng dự án."
                ;;
            7)
                print_info "Đang hiển thị logs lỗi (Ctrl+C để thoát)..."
                print_info "Chỉ hiển thị logs có level ERROR, WARN, PANIC, hoặc FAILED"
                echo ""
                print_info "Đang lọc logs từ tất cả services..."
                echo ""
                
                # Filter logs for errors, warnings, panics, and failures
                # Use --tail=100 to show recent logs first, then follow
                docker compose logs --tail=100 -f 2>&1 | grep --line-buffered -iE "(error|warn|panic|fatal|failed|exception|erro|❌|⚠)" --color=always || {
                    echo ""
                    print_info "Không tìm thấy logs lỗi trong 100 dòng gần nhất."
                    print_info "Tất cả services có vẻ đang hoạt động bình thường."
                    echo ""
                    print_info "Nếu muốn xem full logs, sử dụng lệnh:"
                    echo "  ${GREEN}docker compose logs -f${NC}"
                    echo ""
                    print_info "Hoặc xem logs của service cụ thể:"
                    echo "  ${GREEN}docker compose logs -f <service_name>${NC}"
                }
                ;;
            8)
                print_info "Đang tạo admin mặc định (DEV/DEMO)..."
                create_default_admin
                ;;
            9)
                print_info "Tạm biệt!"
                exit 0
                ;;
            *)
                print_warning "Lựa chọn không hợp lệ. Vui lòng thử lại."
                ;;
        esac
        
        echo ""
        print_info "Nhấn Enter để quay lại menu chính..."
        read
    done
    
    return 0

# Refresh PATH and GOPATH after Go installation
refresh_go_path() {
    # Add Go to PATH if not already there
    if [[ ":$PATH:" != *":/usr/local/go/bin:"* ]]; then
        export PATH=$PATH:/usr/local/go/bin
    fi
    
    # Set GOPATH if not set
    if [ -z "$GOPATH" ]; then
        export GOPATH=$HOME/go
        if ! grep -q 'export GOPATH=' ~/.bashrc 2>/dev/null; then
            echo 'export GOPATH=$HOME/go' >> ~/.bashrc
        fi
    fi
    
    # Add GOPATH/bin to PATH if not already there
    if [[ ":$PATH:" != *":$GOPATH/bin:"* ]] && [[ ":$PATH:" != *":$HOME/go/bin:"* ]]; then
        export PATH=$PATH:$GOPATH/bin
        if ! grep -q '$GOPATH/bin' ~/.bashrc 2>/dev/null; then
            echo 'export PATH=$PATH:$GOPATH/bin' >> ~/.bashrc
        fi
    fi
    
    # Create GOPATH directories if they don't exist
    mkdir -p $GOPATH/bin
}

# Installation functions
install_go() {
    # Always use required version for project compatibility
    local version_to_install=$GO_VERSION_REQUIRED
    print_info "Installing Go $version_to_install (phiên bản yêu cầu của dự án)..."
    
    if [[ "$OS" == "linux" ]] || [[ "$OS" == "wsl" ]]; then
        # Remove old Go installation if exists
        if [ -d "/usr/local/go" ]; then
            print_warning "Removing old Go installation..."
            sudo rm -rf /usr/local/go
        fi
        
        # Download and install Go
        cd /tmp
        wget -q "https://go.dev/dl/go${version_to_install}.linux-amd64.tar.gz"
        if [ $? -eq 0 ]; then
            sudo tar -C /usr/local -xzf "go${version_to_install}.linux-amd64.tar.gz"
            rm "go${version_to_install}.linux-amd64.tar.gz"
        else
            print_error "Failed to download Go. Please check your internet connection."
            return 1
        fi
        
        # Add to PATH
        if ! grep -q '/usr/local/go/bin' ~/.bashrc; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.bashrc
        fi
        
        # Refresh PATH immediately
        refresh_go_path
        
    elif [[ "$OS" == "macos" ]]; then
        if check_command brew; then
            brew install go@${version_to_install} || brew install go
        else
            # Manual installation for macOS
            cd /tmp
            wget -q "https://go.dev/dl/go${version_to_install}.darwin-amd64.tar.gz"
            if [ $? -eq 0 ]; then
                sudo rm -rf /usr/local/go
                sudo tar -C /usr/local -xzf "go${version_to_install}.darwin-amd64.tar.gz"
                rm "go${version_to_install}.darwin-amd64.tar.gz"
            else
                print_error "Failed to download Go. Please check your internet connection."
                return 1
            fi
            
            if ! grep -q '/usr/local/go/bin' ~/.zshrc; then
                echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.zshrc
            fi
        fi
        
        # Refresh PATH immediately
        refresh_go_path
    fi
    
    # Verify Go installation
    if check_command go; then
        local go_version=$(go version | awk '{print $3}' | sed 's/go//')
        print_success "Go installed successfully: $go_version"
    else
        print_error "Go installation completed but 'go' command not found. Please restart your terminal."
        return 1
    fi
}

install_nodejs() {
    # Always use required version for project compatibility
    local version_to_install=$NODE_VERSION_REQUIRED
    print_info "Installing Node.js v${version_to_install} (LTS - phiên bản yêu cầu của dự án)..."
    
    if [[ "$OS" == "linux" ]] || [[ "$OS" == "wsl" ]]; then
        # Install Node.js using NodeSource repository (always use required version)
        curl -fsSL https://deb.nodesource.com/setup_${version_to_install}.x | sudo -E bash -
        sudo apt-get install -y nodejs
        
    elif [[ "$OS" == "macos" ]]; then
        if check_command brew; then
            brew install node@${version_to_install} || brew install node
        else
            print_error "Homebrew not found. Please install Homebrew first: https://brew.sh"
            exit 1
        fi
    fi
    
    print_success "Node.js and npm installed successfully"
}

install_docker() {
    print_info "Installing Docker..."
    
    if [[ "$OS" == "linux" ]] || [[ "$OS" == "wsl" ]]; then
        # Install Docker using official script
        curl -fsSL https://get.docker.com -o get-docker.sh
        sudo sh get-docker.sh
        rm get-docker.sh
        
        # Add user to docker group (requires logout/login)
        sudo usermod -aG docker $USER
        print_warning "User added to docker group. You may need to logout and login again."
        
    elif [[ "$OS" == "macos" ]]; then
        print_info "Please install Docker Desktop for macOS: https://www.docker.com/products/docker-desktop"
        print_warning "Docker Desktop installation requires manual steps."
        read -p "Press Enter after installing Docker Desktop..."
    fi
    
    install_docker_compose
    print_success "Docker installed successfully"
}

install_docker_compose() {
    print_info "Installing Docker Compose..."
    
    if [[ "$OS" == "linux" ]] || [[ "$OS" == "wsl" ]]; then
        # Docker Compose V2 is included with Docker Desktop
        # For standalone installation:
        if ! docker compose version &> /dev/null; then
            DOCKER_COMPOSE_VERSION="v2.24.0"
            sudo curl -L "https://github.com/docker/compose/releases/download/${DOCKER_COMPOSE_VERSION}/docker-compose-linux-x86_64" -o /usr/local/bin/docker-compose
            sudo chmod +x /usr/local/bin/docker-compose
        fi
    fi
    
    print_success "Docker Compose installed successfully"
}

install_sqlc() {
    # Always use required version for project compatibility
    local version_to_install=$SQLC_VERSION_REQUIRED
    print_info "Installing sqlc $version_to_install (phiên bản yêu cầu của dự án)..."
    
    # Refresh PATH before checking
    refresh_go_path
    
    if check_command go; then
        go install github.com/sqlc-dev/sqlc/cmd/sqlc@${version_to_install}
        
        if [ $? -eq 0 ]; then
            print_success "sqlc installed successfully"
        else
            print_error "Failed to install sqlc. Please check your internet connection."
            return 1
        fi
    else
        print_error "Go is required to install sqlc. Please install Go first."
        return 1
    fi
}

install_air() {
    # Always use required version for project compatibility
    local version_to_install=$AIR_VERSION_REQUIRED
    print_info "Installing air $version_to_install (phiên bản yêu cầu của dự án)..."
    
    # Refresh PATH before checking
    refresh_go_path
    
    if check_command go; then
        go install github.com/cosmtrek/air@${version_to_install}
        
        if [ $? -eq 0 ]; then
            print_success "air installed successfully"
        else
            print_error "Failed to install air. Please check your internet connection."
            return 1
        fi
    else
        print_error "Go is required to install air. Please install Go first."
        return 1
    fi
}

install_golangci_lint() {
    # Always use required version for project compatibility
    local version_to_install=$GOLANGCI_LINT_VERSION_REQUIRED
    print_info "Installing golangci-lint $version_to_install (phiên bản yêu cầu của dự án)..."
    
    # Refresh PATH before checking
    refresh_go_path
    
    if ! check_command go; then
        print_error "Go is required to install golangci-lint. Please install Go first."
        return 1
    fi
    
    # Get GOPATH
    local go_path=$(go env GOPATH 2>/dev/null)
    if [ -z "$go_path" ]; then
        go_path=$HOME/go
    fi
    
    local install_dir="$go_path/bin"
    
    # Create bin directory if it doesn't exist
    mkdir -p "$install_dir"
    
    if [[ "$OS" == "linux" ]] || [[ "$OS" == "wsl" ]]; then
        curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$install_dir" ${version_to_install}
        
        if [ $? -eq 0 ]; then
            print_success "golangci-lint installed successfully to $install_dir"
        else
            print_error "Failed to install golangci-lint. Please check your internet connection."
            return 1
        fi
        
    elif [[ "$OS" == "macos" ]]; then
        if check_command brew; then
            brew install golangci-lint
        else
            curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b "$install_dir" ${version_to_install}
        fi
    fi
}

# Run main function
main "$@"

