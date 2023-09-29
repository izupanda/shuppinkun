from flask import Flask, request, jsonify
from scrape_amazon import scrape_amazon  # スクレイピングの関数をインポート

app = Flask(__name__)

@app.route('/get-price', methods=['GET'])
def get_price():
    
    query = request.args.get('query')  # クエリパラメータから検索文字列を取得
    if not query:
        return jsonify({"error": "Query parameter is missing"}), 400

    result = scrape_amazon(query)
    if result:

        return jsonify(result)
    else:
        return jsonify({"error": "Could not fetch data"}), 500

if __name__ == '__main__':
    app.run(debug=True)
