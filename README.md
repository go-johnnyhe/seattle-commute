# Seattle Commute CLI

A simple command-line tool for Seattle commuters to get optimal transit routes and schedules using Google Maps transit data.

## Features

- **Zero-friction UX**: No location detection needed - assumes you're commuting between home and work
- Routes home with `commute` (assumes you're at work)
- Routes to work with `commute -w` (assumes you're at home)
- **Arbitrary routing** with `commute "from" "to"` for any Seattle locations
- Shows next several departure times in the next 2 hours
- Real-time transit information for Seattle Metro, Sound Transit, and more
- Saves your home/work addresses for quick access

## Quick Start

1. **Get a Google Maps API Key**:
   - Go to [Google Cloud Console](https://console.cloud.google.com/)
   - Enable the **Directions API**
   - Create an API key
   - (Free tier includes 40,000 requests/month, you will not run out of this LOL) 

2. **Install and Setup**:
   ```bash
   go build -o commute
   ./commute init "123 Your St, Seattle, WA"
   ```

3. **Daily Usage**:
   ```bash
   ./commute                              # Routes home (assumes you're at work)
   ./commute -w                           # Routes to work (assumes you're at home)
   ./commute "U District" "Capitol Hill"  # Arbitrary routing between any two places
   ```

## Commands

### `commute init [address]`
Set up your home address and Google Maps API key.

**Interactive setup**:
```bash
./commute init
```

**Quick setup**:
```bash
./commute init "123 Main St, Seattle, WA"
```

### `commute`
Get transit routes home (assumes you're at work for zero-friction UX).

**Example output**:
```
📍 Going home (assuming you're at work)
📏 Checking distance... ✅
🚌 Finding transit routes... ✅

🏠 Routes to 123 Main St, Seattle, WA (home)
=====================================

1. Depart: 2:15 PM (5m) ⚡ GOOD TIMING
   Arrive: 2:47 PM (Travel: 32m)
   Distance: 8.2 mi
   🚌 Bus 40 → Link Light Rail

2. Depart: 2:28 PM (18m)
   Arrive: 3:05 PM (Travel: 37m)
   Distance: 9.1 mi
   🚌 Bus 150 → Bus 21

📱 Tip: Add this tool to your PATH for quick access anywhere!
```

### `commute -w`
Get transit routes to work (assumes you're at home for zero-friction UX).

### `commute "from" "to"`
Get transit routes between any two arbitrary locations in Seattle.

**Examples**:
```bash
./commute "U District Station" "Capitol Hill"
./commute "Pike Place Market" "Space Needle"
./commute "Redmond Transit Center" "University of Washington"
```

## Configuration

Config is stored at `~/.seattle-commute/config.json`:

```json
{
  "home_address": "123 Main St, Seattle, WA",
  "work_address": "456 Work Ave, Seattle, WA",
  "google_api_key": "your-api-key-here"
}
```

## Installation

**Option 1: Build from source**
```bash
git clone <this-repo>
cd seattle-commute-cli
go build -o commute
```

**Option 2: Add to PATH for global access**
```bash
go build -o commute
sudo mv commute /usr/local/bin/
# Now you can run 'commute' from anywhere!
```

## How It Works

1. **Zero-Friction UX**: Assumes you're commuting between home and work (no location detection needed)
2. **Transit API**: Queries Google Maps Directions API with transit mode
3. **Smart Scheduling**: Gets multiple departure times over the next 2 hours
4. **Route Optimization**: Shows the best routes sorted by departure time
5. **Real-time Data**: Includes live Seattle Metro, Sound Transit, and streetcar schedules

## Requirements

- Go 1.18+
- Google Maps API key with Directions API enabled
- Internet connection for real-time transit data

## Privacy & Data

- Your location is detected via IP address only
- No location data is stored or transmitted except to Google Maps API
- Config file contains only addresses you provide and your API key
- All data stays on your local machine

## Troubleshooting

**"No routes found"**:
- Check that your addresses are valid Seattle-area locations
- Verify transit service is available at the current time

**"Configuration not found"**:
- Run `commute init` to set up your addresses and API key

**"Failed to get location"** (only when using `--from` flag):
- Check your internet connection
- IP geolocation fallback may be affected by VPNs

## Contributing

This tool was designed for Seattle commuters. Pull requests welcome for:
- Better error handling
- Additional output formats
- Performance improvements
- Bug fixes

## License

MIT License - feel free to modify and distribute!
