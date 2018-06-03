from bs4 import BeautifulSoup
import sys
import re
import requests
import lxml
def main():
    r=requests.get(sys.argv[1])
    soup = BeautifulSoup(r.text,"lxml")
    divs = soup.find(id="docs-content").find_all("div",{'class': 'section'})
    content = divs[0].get_text()
    newcontent = re.sub('(\n)+', '\n', content)
    print(newcontent)
