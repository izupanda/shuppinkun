# app.py
from flask import Flask, jsonify
from scrape_amazon import scrape_amazon  # ここでスクレイピング関数をインポート

app = Flask(__name__)

@app.route('/get-price')
def get_price():
    # Amazonからデータをスクレイピング
    url = "https://www.amazon.co.jp/some-product-url"  # 実際のAmazonの商品URLを設定してください。
    result = scrape_amazon(url)  # スクレイピング関数を呼び出して結果を取得

    if result:  # スクレイピングが成功した場合
        print("Sending response:", result)
        return jsonify(result)
    else:  # スクレイピングが失敗した場合
        return jsonify({"error": "Failed to scrape data"}), 500

if __name__ == '__main__':
    app.run(debug=True)
