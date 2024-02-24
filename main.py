import requests
from bs4 import BeautifulSoup

def ygl_listings(url: str):
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

for listing in ygl_listings('https://ygl.is/99333?beds_from=4&beds_to=5&rent_to=5200&date_from=09%2F01%2F2024'):
    print(listing.a['href'])
