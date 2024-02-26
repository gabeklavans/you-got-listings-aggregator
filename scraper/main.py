#!/usr/bin/env python3

import json
import argparse
from typing import Dict
from dotenv import load_dotenv
import requests
from bs4 import BeautifulSoup
import bot

load_dotenv()

parser = argparse.ArgumentParser()
parser.add_argument('--notify', action='store_true', help='Enable notifications for new listings')
args = parser.parse_args()

def ygl_listings(url: str):
    '''Generator function for getting all paginated listings from a ygl search query'''
    page = 0

    while True:
        page += 1
        response = requests.get(f'{url}&page={page}', timeout=10)
        response.raise_for_status()
        html_content = response.text

        soup = BeautifulSoup(html_content, 'lxml')

        if soup.find('div', class_='nothing_found') is not None:
            break

        listings = soup.find_all('div', class_='property_item')
        for listing in listings:
            yield listing

def fill_properties(listings: Dict, ygl_url_base: str):
    '''
    Fill a persistent props dict with listings and their data

    json format example
    {
        "100 Beefcake Rd": {
            "refs": ["ygl.is/12345/678910", "ygl.is/12/34"],
            "price": 4400,
            "beds": 4,
            "baths": 2,
            "date": "09/01/2024",
        }
    }
    '''
    for listing in ygl_listings(f'{ygl_url_base}?beds_from=4&beds_to=5&rent_to=5200&date_from=09%2F01%2F2024'):
        listing_element = listing.find('a', class_='item_title')
        listing_addr = listing_element.get_text()
        listing_url = listing_element['href']

        if listing_addr not in listings:
            # initialize a new entry for this listing
            if args.notify:
                bot.notify(listing_url)

            listings[listing_addr] = {
                'refs': []
            }

            listing_props_elements = listing.find_all('div', class_='column')
            listing_props = list(map(lambda tag: tag.text.strip(), listing_props_elements))
            # the listing properties are well-ordered, so we parse them directly
            listing_price = int(''.join(filter(lambda char: char.isdigit(), listing_props[0])))
            listing_beds = float(listing_props[1].split(' ')[0])
            listing_baths = float(listing_props[2].split(' ')[0])
            listing_date = listing_props[3].split(' ')[1]

            listings[listing_addr]['price'] = listing_price
            listings[listing_addr]['beds'] = listing_beds
            listings[listing_addr]['baths'] = listing_baths
            listings[listing_addr]['date'] = listing_date

        # always check if this is a new copy of the listing
        if listing_url not in listings[listing_addr]['refs']:
            listings[listing_addr]['refs'].append(listing_url)


if __name__ == "__main__":
    with open('../sites.json', 'r', encoding='utf-8') as sites_fp:
        sites = json.load(sites_fp)

    try:
        with open('listings.json', 'r', encoding='utf-8') as listings_fp:
            total_props = json.load(listings_fp)
    except IOError as e:
        total_props = {}

    for site in sites.keys():
        fill_properties(total_props, site)

    with open('listings.json', 'w', encoding='utf-8') as listings_file:
        json.dump(total_props, listings_file)
