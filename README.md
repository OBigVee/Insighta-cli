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
