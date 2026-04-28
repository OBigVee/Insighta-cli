# Insighta Labs+ CLI

The official Command Line Interface for the [Insighta Labs+](https://github.com/OBigVee/data-and-API) Profile Intelligence Platform.

## Features
- **Secure Authentication**: Uses GitHub OAuth with PKCE flow.
- **Automated Token Management**: Automatically refreshes expired JWTs in the background.
- **Role-Based Access**: Respects server-side Admin/Analyst roles.
- **Rich Display**: Outputs clean, box-drawn ASCII tables with loading spinners.

## Installation

```bash
# Clone the repository
git clone <your-cli-repo-url>
cd insighta-cli

# Build the executable
go build -o insighta .

# (Optional) Move to your PATH
sudo mv insighta /usr/local/bin/
```

### Docker

You can also run the CLI as a container:

```bash
# Build the image
docker build -t insighta-cli .

# Run a command (e.g., list profiles)
# Note: you can mount a volume to persist your login credentials
docker run -it -v ~/.insighta:/root/.insighta insighta-cli profiles list
```

### Docker Compose (Recommended for sharing)

If you have Docker Compose installed, you can run commands even more easily:

```bash
# Build and run a command
docker-compose run --rm cli profiles list
```

This method automatically handles volume mounting and interactive terminal settings for you.

## Usage

### Authentication
Authenticate with your GitHub account:
```bash
./insighta login
```
Check your current logged-in status:
```bash
./insighta whoami
```

### Profile Management

**List Profiles:**
```bash
./insighta profiles list
```
*Optional Flags:*
- `--gender` (male/female)
- `--country` (country code, e.g., NG)
- `--min-age` / `--max-age`
- `--age-group` (child, teenager, adult, senior)
- `--sort-by` (age, gender_probability, created_at)
- `--order` (asc, desc)
- `--page` / `--limit`

**Natural Language Search:**
```bash
./insighta profiles search "adult women from Nigeria"
```

**Get Profile Details:**
```bash
./insighta profiles get <profile_id>
```

**Export Profiles (CSV):**
```bash
./insighta profiles export --format csv --gender female
```

**Create Profile (Admin Only):**
```bash
./insighta profiles create --name "John Doe"
```

## Configuration
Authentication credentials and tokens are stored securely in `~/.insighta/credentials.json`. To clear them manually, run `./insighta logout`.
