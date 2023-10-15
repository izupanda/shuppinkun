from bs4 import BeautifulSoup
import requests
import json
import sys
import traceback

def scrape_amazon(search_query):
    try:
        base_url = "https://www.amazon.co.jp/s?k="
        search_url = f"{base_url}{search_query}"
        headers = {
            "User-Agent": "Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/58.0.3029.110 Safari/537"
        }

        response = requests.get(search_url, headers=headers)

        if response.status_code != 200:
            print(f"Failed to get the webpage, status code: {response.status_code}", file=sys.stderr)
            return None

        soup = BeautifulSoup(response.content, "html.parser")
        results = soup.find_all("div", {"data-component-type": "s-search-result"})

        for result in results:
            title = result.find("span", class_="a-text-normal")
            price = result.find("span", class_="a-price-whole")
            image = result.find("img", class_="s-image")

            if title and price and image:
                data = {
                    "title": title.text,
                    "price": f"¥{price.text}",
                    "image_url": image["src"]
                }
                return data

        # If we reach here, it means we couldn't find a product match
        print("Could not find product details.", file=sys.stderr)
        return None
    except Exception as e:
        print(f"Error occurred: {e}", file=sys.stderr)
        traceback.print_exc(file=sys.stderr)  # これを追加します
        return None

if __name__ == "__main__":
    # コマンドライン引数が提供されているか確認します
    if len(sys.argv) > 1:
        search_query = sys.argv[1]  # 最初のコマンドライン引数を取得
    else:
        print("Please provide a search query as an argument.")
        sys.exit(1)  # スクリプトを終了
    
    result = scrape_amazon(search_query)
    if result:
        print(json.dumps(result))
    else:
        print("No result found.")
