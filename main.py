#!/usr/bin/env python3

import json
from typing import Dict
import requests
from bs4 import BeautifulSoup

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


def fill_properties(props: Dict, ygl_url_base: str):
    '''
    Fill a persistent props dict with listings and their data

    json format example
    {
        '100 Beefcake Rd': {
            'refs': ['ygl.is/12345/678910', 'ygl.is/12/34']
        }
    }
    '''
    for listing in ygl_listings(f'{ygl_url_base}?beds_from=4&beds_to=5&rent_to=5200&date_from=09%2F01%2F2024'):
        prop_element = listing.find('a', class_='item_title')
        prop_addr = prop_element.get_text()
        prop_url = prop_element['href']
        if prop_addr not in props:
            props[prop_addr] = { 'refs': [] }
            print(f'New listing {prop_addr} found!')
        if prop_url not in props[prop_addr]['refs']:
            props[prop_addr]['refs'].append(prop_url)


if __name__ == "__main__":
    try:
        with open('listings.json', 'r', encoding='utf-8') as listings_fp:
            total_props = json.load(listings_fp)
    except IOError as e:
        total_props = {}

    fill_properties(total_props, 'https://ygl.is/99333')
    fill_properties(total_props, 'https://ygl.is/99334')

    with open('listings.json', 'w', encoding='utf-8') as listings_file:
        json.dump(total_props, listings_file)
