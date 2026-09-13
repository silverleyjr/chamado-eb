#!/usr/bin/env bash
# One-time setup for a fresh Linux Mint / Ubuntu-based machine: installs
# everything ./start.sh needs (Docker, Go, Python3) and adds the current
# user to the docker group. Safe to re-run; skips what's already installed.
set -euo pipefail

if ! command -v apt-get >/dev/null 2>&1; then
	echo "This script targets Debian/Ubuntu-based systems (like Linux Mint)." >&2
	echo "apt-get was not found — install Docker, Go and Python3 manually for your distro." >&2
	exit 1
fi

echo "==> Updating package lists..."
sudo apt-get update

echo "==> Installing base tools (curl, ca-certificates, python3)..."
sudo apt-get install -y curl ca-certificates python3

echo "==> Installing Docker Engine + Compose plugin..."
if command -v docker >/dev/null 2>&1; then
	echo "Docker already installed ($(docker --version)), skipping."
else
	sudo apt-get install -y docker.io docker-compose-v2
fi

echo "==> Enabling the Docker service..."
sudo systemctl enable --now docker

NEEDS_RELOGIN=0
echo "==> Adding $USER to the docker group..."
if groups "$USER" | grep -qw docker; then
	echo "$USER is already in the docker group."
else
	sudo usermod -aG docker "$USER"
	NEEDS_RELOGIN=1
fi

echo "==> Installing Go..."
if command -v go >/dev/null 2>&1; then
	echo "Go already installed: $(go version)"
else
	case "$(uname -m)" in
		x86_64)  GOARCH=amd64 ;;
		aarch64) GOARCH=arm64 ;;
		*) echo "Unsupported architecture: $(uname -m). Install Go manually from https://go.dev/dl/" >&2; exit 1 ;;
	esac

	GO_VERSION="$(curl -fsSL https://go.dev/VERSION?m=text | head -1)"
	echo "Downloading ${GO_VERSION} for linux-${GOARCH}..."
	curl -fsSL "https://go.dev/dl/${GO_VERSION}.linux-${GOARCH}.tar.gz" -o /tmp/go.tar.gz
	sudo rm -rf /usr/local/go
	sudo tar -C /usr/local -xzf /tmp/go.tar.gz
	rm /tmp/go.tar.gz

	if ! grep -qs '/usr/local/go/bin' "$HOME/.profile"; then
		echo 'export PATH=$PATH:/usr/local/go/bin' >> "$HOME/.profile"
	fi
	export PATH="$PATH:/usr/local/go/bin"
fi

echo
echo "======================================================"
echo " Install complete."
echo "   Go:      $(command -v go >/dev/null 2>&1 && go version || echo 'installed to /usr/local/go — open a new shell')"
echo "   Docker:  $(docker --version)"
echo "======================================================"
if [ "$NEEDS_RELOGIN" = "1" ]; then
	echo
	echo " IMPORTANT: you were just added to the docker group. Log out and back"
	echo " in (or run: newgrp docker) before running ./start.sh, otherwise Docker"
	echo " commands will fail with a permission error."
else
	echo
	echo " Next: run ./start.sh to launch Chamado."
fi
