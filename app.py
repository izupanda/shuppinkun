from flask import Flask, request, jsonify
import json
from scrape_amazon import scrape_amazon

app = Flask(__name__)

@app.route('/get-price', methods=['GET'])
def get_price():
    try:
        query = request.args.get('query')
        print(f"Received query: {query}")

        if not query:
            return jsonify({"error": "Query parameter is missing"}), 400

        result = scrape_amazon(query)

        if not result:
            # 詳細なエラーメッセージを追加
            return jsonify({
                "error": "Could not fetch data",
                "detail": f"Failed to retrieve data for query: {query}"
            }), 500

        print(f"Before jsonify: {result}")

        response = jsonify(result)
        response.headers['Content-Type'] = 'application/json; charset=utf-8'
        return response
    except Exception as e:
        print(f"An exception occurred: {e}")
        return jsonify({"error": str(e)}), 500

if __name__ == '__main__':
    app.run(debug=True)
