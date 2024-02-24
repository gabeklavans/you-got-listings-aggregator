import requests
from bs4 import BeautifulSoup

class YGL:
    def __init__(self, ygl_url: str) -> None:
        self.ygl_url = ygl_url
        self.cur_page_properties = []
        self.cur_page_idx = 0
        self.page = 0

    def __iter__(self):
        return self

    def __next__(self):
        if self.cur_page_idx >= len(self.cur_page_properties):
            # get the next page of results
            self.page += 1
            url = f'{self.ygl_url}&page={self.page}'
            response = requests.get(url, timeout=10)
            response.raise_for_status()
            html_content = response.text

            soup = BeautifulSoup(html_content, 'lxml')

            if soup.find('div', class_='nothing_found') is not None:
                raise StopIteration

            self.cur_page_properties = soup.find_all('div', class_='property_item')

            self.cur_page_idx = 0

        prop = self.cur_page_properties[self.cur_page_idx]
        self.cur_page_idx += 1
        return prop

for ygl_prop in YGL('https://ygl.is/99333?beds_from=4&beds_to=5&rent_to=5200&date_from=09%2F01%2F2024'):
    print(ygl_prop.a['href'])
    # pass
