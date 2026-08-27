package config

const DefaultConfig = `# Ritta deployment configuration

# Local project directory
local_project_root: ./

# Directory where the project will be deployed
remote_project_root: ~/srv/

setup_config:
  # Setup script to run on the server
  # You can modify this, default is set up with apt package manager and default installs
  script: ./rittaScript.sh

source:
  # "existing" = use the current project(the project is already on your vps)
  # "git"      = clone from a repository
  type: existing

  # Uncomment these when using git:
  # repository: git@github.com:you/project.git
  # branch: main


  # Vps info and your ssh key for log in
server:
  host: ""
  user: deploy
  port: 22
  key: ~/.ssh/id_ed25519


  # Optional health check command if you have
health:
  command: ""


# Files to copy to the server
# Add as many files as needed
file:
  # - from: .env
  #   to: .env

  # - from: ./config/app.yaml
  #   to: config/app.yaml


  # Your application specific build command 
build:
  command: ""

# Your application specific run command
run:
  command: ""


  # Domains to expose through the reverse proxy(optional)
# Add as many domains as needed
domains:
  # - host: api.example.com
  #   port: 8000
  #   tls: true

  # - host: example.com
  #   port: 3000
  #   tls: true

proxy:
  provider: Nginx

tls:
  provider: letsencrypt
  email: ""
`