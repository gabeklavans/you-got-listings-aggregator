# Dev Folder
Contains useful scripts and files for testing. Feel free to add your own.

## Permissions
You'll need to add execute permissions for every script in this folder. The easiest way to do this is to run this command:

`chmod +x ./dev/*`

You'll have to do this for every script you add.

## Scripts
### add_filters.sh
- Populates `Filter` table of ygl.db. Useful for quick repopulation.
- Usage: `./dev/add_filters.sh`

### clear_table.sh
- Removes all rows of a table. Takes in name of a table as an argument.
- Usage: `./dev/clear_table.sh {table_name}`
- Example: `./dev/clear_table.sh Filter`

### clear_brokers.sh
- Removes all rows of the `Broker` table of ygl.db. 
- Usage: `./dev/clear_brokers.sh`
- Calls: [clear_table.sh](#clear_tablesh)

### clear_filters.sh
- Removes all rows of the `Filter` table of ygl.db.
- Usage: `./dev/clear_filters.sh`
- Calls: [clear_table.sh](#clear_tablesh)

### clear_listings.sh
- Removes all rows of the `Listing` table of ygl.db.
- Usage: `./dev/clear_listings.sh`
- Calls: [clear_table.sh](#clear_tablesh)

### run_scraper.sh
- Runs the web scraper, adding relevant listings to the `Listing` table.
- Usage: `./dev/run_scraper.sh`
- Calls: [scraper/main.py](/scraper/main.py)

### refresh_listings.sh
- Clears all existing listings and runs the scraper.
- Usage: `./dev/refresh_listings.sh`
- Calls: [clear_listings.sh](#clear_listingssh), [run_scraper.sh](#run_scrapersh)


