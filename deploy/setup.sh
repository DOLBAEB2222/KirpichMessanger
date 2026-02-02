#!/bin/bash

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
MIN_RAM_GB=3
MIN_DISK_GB=15
REQUIRED_UBUNTU_VERSION="24.04"

echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}Messenger Application Setup Script${NC}"
echo -e "${GREEN}Ubuntu 24.04 LTS - Optimized Deployment${NC}"
echo -e "${GREEN}========================================${NC}\n"

# Check if running as root
if [[ $EUID -eq 0 ]]; then
   echo -e "${RED}This script should NOT be run as root${NC}"
   echo -e "Please run as a regular user with sudo privileges"
   exit 1
fi

# Check Ubuntu version
check_ubuntu_version() {
    echo -e "${YELLOW}[1/10]${NC} Checking Ubuntu version..."
    
    if [ -f /etc/os-release ]; then
        . /etc/os-release
        if [[ "$VERSION_ID" != "$REQUIRED_UBUNTU_VERSION" ]]; then
            echo -e "${RED}Warning: This script is designed for Ubuntu ${REQUIRED_UBUNTU_VERSION}${NC}"
            echo -e "Current version: $VERSION_ID"
            read -p "Continue anyway? (y/N) " -n 1 -r
            echo
            if [[ ! $REPLY =~ ^[Yy]$ ]]; then
                exit 1
            fi
        else
            echo -e "${GREEN}✓ Ubuntu ${VERSION_ID} detected${NC}"
        fi
    fi
}

# Check system requirements
check_requirements() {
    echo -e "\n${YELLOW}[2/10]${NC} Checking system requirements..."
    
    # Check RAM
    total_ram_gb=$(free -g | awk '/^Mem:/{print $2}')
    if [ "$total_ram_gb" -lt "$MIN_RAM_GB" ]; then
        echo -e "${RED}✗ Insufficient RAM: ${total_ram_gb}GB (minimum ${MIN_RAM_GB}GB required)${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ RAM: ${total_ram_gb}GB${NC}"
    fi
    
    # Check disk space
    available_disk_gb=$(df -BG / | awk 'NR==2 {print $4}' | sed 's/G//')
    if [ "$available_disk_gb" -lt "$MIN_DISK_GB" ]; then
        echo -e "${RED}✗ Insufficient disk space: ${available_disk_gb}GB (minimum ${MIN_DISK_GB}GB required)${NC}"
        exit 1
    else
        echo -e "${GREEN}✓ Disk space: ${available_disk_gb}GB available${NC}"
    fi
    
    # Check CPU cores
    cpu_cores=$(nproc)
    echo -e "${GREEN}✓ CPU cores: ${cpu_cores}${NC}"
}

# Update system
update_system() {
    echo -e "\n${YELLOW}[3/10]${NC} Updating system packages..."
    sudo apt update
    sudo apt upgrade -y
    echo -e "${GREEN}✓ System updated${NC}"
}

# Install Docker
install_docker() {
    echo -e "\n${YELLOW}[4/10]${NC} Installing Docker..."
    
    if command -v docker &> /dev/null; then
        echo -e "${GREEN}✓ Docker already installed${NC}"
        docker --version
    else
        # Install dependencies
        sudo apt install -y apt-transport-https ca-certificates curl software-properties-common
        
        # Add Docker's official GPG key
        curl -fsSL https://download.docker.com/linux/ubuntu/gpg | sudo gpg --dearmor -o /usr/share/keyrings/docker-archive-keyring.gpg
        
        # Add Docker repository
        echo "deb [arch=$(dpkg --print-architecture) signed-by=/usr/share/keyrings/docker-archive-keyring.gpg] https://download.docker.com/linux/ubuntu $(lsb_release -cs) stable" | sudo tee /etc/apt/sources.list.d/docker.list > /dev/null
        
        # Install Docker
        sudo apt update
        sudo apt install -y docker-ce docker-ce-cli containerd.io docker-compose-plugin
        
        # Add current user to docker group
        sudo usermod -aG docker $USER
        
        # Enable Docker service
        sudo systemctl enable docker
        sudo systemctl start docker
        
        echo -e "${GREEN}✓ Docker installed${NC}"
    fi
}

# Install Go
install_go() {
    echo -e "\n${YELLOW}[5/10]${NC} Installing Go..."
    
    GO_VERSION="1.21.6"
    
    if command -v go &> /dev/null; then
        echo -e "${GREEN}✓ Go already installed${NC}"
        go version
    else
        cd /tmp
        wget -q https://go.dev/dl/go${GO_VERSION}.linux-amd64.tar.gz
        sudo rm -rf /usr/local/go
        sudo tar -C /usr/local -xzf go${GO_VERSION}.linux-amd64.tar.gz
        
        # Add Go to PATH for current user
        if ! grep -q "/usr/local/go/bin" ~/.profile; then
            echo 'export PATH=$PATH:/usr/local/go/bin' >> ~/.profile
        fi
        
        export PATH=$PATH:/usr/local/go/bin
        
        echo -e "${GREEN}✓ Go ${GO_VERSION} installed${NC}"
    fi
}

# Install additional tools
install_tools() {
    echo -e "\n${YELLOW}[6/10]${NC} Installing additional tools..."
    
    sudo apt install -y \
        postgresql-client \
        redis-tools \
        git \
        wget \
        curl \
        htop \
        net-tools \
        ufw \
        fail2ban
    
    echo -e "${GREEN}✓ Additional tools installed${NC}"
}

# Configure firewall
configure_firewall() {
    echo -e "\n${YELLOW}[7/10]${NC} Configuring firewall..."
    
    # Check if UFW is active
    if sudo ufw status | grep -q "Status: active"; then
        echo -e "${YELLOW}UFW is already active${NC}"
    else
        # Configure UFW
        sudo ufw default deny incoming
        sudo ufw default allow outgoing
        
        # Allow SSH
        sudo ufw allow 22/tcp
        
        # Allow HTTP/HTTPS
        sudo ufw allow 80/tcp
        sudo ufw allow 443/tcp
        sudo ufw allow 443/udp
        
        # Allow TURN server ports
        sudo ufw allow 3478:3479/tcp
        sudo ufw allow 3478:3479/udp
        sudo ufw allow 49152:49252/udp
        
        # Enable firewall
        echo "y" | sudo ufw enable
    fi
    
    sudo ufw status
    echo -e "${GREEN}✓ Firewall configured${NC}"
}

# Setup project directory
setup_project() {
    echo -e "\n${YELLOW}[8/10]${NC} Setting up project directory..."
    
    PROJECT_DIR="$HOME/messenger"
    
    if [ -d "$PROJECT_DIR" ]; then
        echo -e "${YELLOW}Project directory already exists at $PROJECT_DIR${NC}"
        read -p "Remove and re-clone? (y/N) " -n 1 -r
        echo
        if [[ $REPLY =~ ^[Yy]$ ]]; then
            rm -rf "$PROJECT_DIR"
        else
            echo -e "${YELLOW}Skipping project setup${NC}"
            return
        fi
    fi
    
    echo -e "${YELLOW}Enter repository URL (or press Enter to skip):${NC}"
    read -r REPO_URL
    
    if [ -n "$REPO_URL" ]; then
        git clone "$REPO_URL" "$PROJECT_DIR"
        cd "$PROJECT_DIR"
    else
        mkdir -p "$PROJECT_DIR"
        cd "$PROJECT_DIR"
        echo -e "${YELLOW}Skipped repository cloning${NC}"
    fi
    
    # Create necessary directories
    mkdir -p data/media
    mkdir -p logs
    mkdir -p backups
    
    echo -e "${GREEN}✓ Project directory setup complete${NC}"
}

# Create environment file
create_env_file() {
    echo -e "\n${YELLOW}[9/10]${NC} Creating environment configuration..."
    
    ENV_FILE="$HOME/messenger/deploy/.env"
    
    if [ -f "$ENV_FILE" ]; then
        echo -e "${YELLOW}.env file already exists${NC}"
        return
    fi
    
    # Generate random secrets
    JWT_SECRET=$(openssl rand -hex 32)
    POSTGRES_PASSWORD=$(openssl rand -hex 16)
    TURN_PASSWORD=$(openssl rand -hex 16)
    
    cat > "$ENV_FILE" << EOF
# Application Configuration
APP_ENV=production
APP_PORT=8080
LOG_LEVEL=info

# Domain Configuration
DOMAIN=localhost
EXTERNAL_IP=auto

# Database Configuration
POSTGRES_DB=messenger
POSTGRES_USER=messenger
POSTGRES_PASSWORD=$POSTGRES_PASSWORD

# Redis Configuration
REDIS_PASSWORD=

# JWT Configuration
JWT_SECRET=$JWT_SECRET
JWT_EXPIRE_HOURS=24

# Upload Configuration
UPLOAD_MAX_SIZE_MB=50

# Rate Limiting
RATE_LIMIT_REQUESTS=100

# CORS Configuration
CORS_ORIGINS=*

# TURN Server Configuration
TURN_USERNAME=turnuser
TURN_PASSWORD=$TURN_PASSWORD
TURN_REALM=messenger.local
EOF
    
    echo -e "${GREEN}✓ Environment file created at $ENV_FILE${NC}"
    echo -e "${YELLOW}⚠ Please review and update the .env file with your actual values!${NC}"
}

# Start services
start_services() {
    echo -e "\n${YELLOW}[10/10]${NC} Starting services..."
    
    cd "$HOME/messenger/deploy"
    
    # Check if docker-compose.yml exists
    if [ ! -f "docker-compose.yml" ]; then
        echo -e "${RED}✗ docker-compose.yml not found${NC}"
        echo -e "${YELLOW}Please ensure the project is properly set up${NC}"
        return
    fi
    
    # Pull images
    echo -e "${YELLOW}Pulling Docker images...${NC}"
    docker compose pull
    
    # Start services
    echo -e "${YELLOW}Starting containers...${NC}"
    docker compose up -d
    
    echo -e "${GREEN}✓ Services started${NC}"
    
    # Wait for services to be healthy
    echo -e "\n${YELLOW}Waiting for services to be healthy...${NC}"
    sleep 10
    
    docker compose ps
}

# Final instructions
print_final_instructions() {
    echo -e "\n${GREEN}========================================${NC}"
    echo -e "${GREEN}Installation Complete!${NC}"
    echo -e "${GREEN}========================================${NC}\n"
    
    echo -e "Next steps:"
    echo -e "1. ${YELLOW}Review configuration:${NC} nano ~/messenger/deploy/.env"
    echo -e "2. ${YELLOW}Update domain:${NC} Set DOMAIN in .env to your actual domain"
    echo -e "3. ${YELLOW}Restart services:${NC} cd ~/messenger/deploy && docker compose restart"
    echo -e "4. ${YELLOW}Check logs:${NC} docker compose logs -f"
    echo -e "5. ${YELLOW}Access API:${NC} http://your-domain/api/v1/health"
    
    echo -e "\n${YELLOW}Useful commands:${NC}"
    echo -e "  docker compose ps              # Check service status"
    echo -e "  docker compose logs -f api     # View API logs"
    echo -e "  docker compose restart         # Restart all services"
    echo -e "  docker compose down            # Stop all services"
    
    echo -e "\n${YELLOW}Database access:${NC}"
    echo -e "  docker exec -it messenger-postgres psql -U messenger -d messenger"
    
    echo -e "\n${RED}Security reminder:${NC}"
    echo -e "  - Change default passwords in .env"
    echo -e "  - Configure SSL certificate for production"
    echo -e "  - Set up regular backups"
    echo -e "  - Review firewall rules"
    
    if ! groups $USER | grep -q docker; then
        echo -e "\n${YELLOW}⚠ Important:${NC} You need to log out and log back in for Docker group changes to take effect"
    fi
}

# Main execution
main() {
    check_ubuntu_version
    check_requirements
    update_system
    install_docker
    install_go
    install_tools
    configure_firewall
    setup_project
    create_env_file
    
    # Only start services if we're in a project directory with docker-compose.yml
    if [ -f "$HOME/messenger/deploy/docker-compose.yml" ]; then
        read -p "Start services now? (Y/n) " -n 1 -r
        echo
        if [[ ! $REPLY =~ ^[Nn]$ ]]; then
            start_services
        fi
    fi
    
    print_final_instructions
}

# Run main function
main
