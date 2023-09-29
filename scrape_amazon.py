from bs4 import BeautifulSoup
import requests
import json
import sys

# スクレイピングのコード
def scrape_amazon(search_query):
    base_url = "https://www.amazon.co.jp/s?k="
    search_url = f"{base_url}{search_query}"

    headers = {
        "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537.3"
    }

    response = requests.get(search_url, headers=headers)

    if response.status_code != 200:
        sys.stderr.write(f"Failed to get the webpage, status code: {response.status_code}\n")
        return

    soup = BeautifulSoup(response.content, "html.parser")

    # Debug: Output the HTML to a file
    with open("debug.html", "w", encoding="utf-8") as f:
        f.write(soup.prettify())

    # スポンサープロダクトを除外
    results = soup.find_all("div", {"data-component-type": "s-search-result"})
    first_non_sponsored_result = None
    for result in results:
        if not result.find("span", class_="s-card-container "):
            first_non_sponsored_result = result
            break

    if not first_non_sponsored_result:
        sys.stderr.write("Could not find a non-sponsored item\n")
        return

    try:
        title = first_non_sponsored_result.find("span", {"class": "a-text-normal"}).text
    except AttributeError:
        sys.stderr.write("Could not find the title element\n")
        return
    
    try:
        price = first_non_sponsored_result.find("span", {"class": "a-price-whole"}).text
    except AttributeError:
        sys.stderr.write("Could not find the price element\n")
        return

    try:
        image_url = first_non_sponsored_result.find("img", {"class": "s-image"})["src"]
    except AttributeError:
        sys.stderr.write("Could not find the image element\n")
        return

    output_data = {
        "title": title,
        "price": f"¥{price}",
        "image_url": image_url
    }

    return output_data  # ここでJSONデータを返します

# 以下の部分はスクリプトとして直接実行された場合にのみ実行されます
if __name__ == "__main__":
    test_search_query = "WF1000XM3"
    result = scrape_amazon(test_search_query)
    if result:
        print(json.dumps(result))
