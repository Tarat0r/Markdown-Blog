import dotenv
from dotenv import load_dotenv
from flask import Flask, render_template, jsonify
import os
import requests
from dotenv import load_dotenv


load_dotenv()
# Настройки скрипта
SERVER_URL = os.getenv("SERVER_URL")  # URL API сервера
API_TOKEN = os.getenv("API_TOKEN")  # Токен API

app = Flask(__name__)


@app.route("/")
def index():
    """Главная страница, загружающая пустой шаблон"""
    return render_template("index.html")


@app.route("/api/notes")
def get_notes():
    """Возвращает список заметок с сервера"""
    headers = {"Authorization": API_TOKEN}
    response = requests.get(f"{SERVER_URL}/notes", headers=headers)

    if response.status_code == 200:
        notes = response.json()
        return jsonify(notes)
    else:
        return jsonify({"error": "Failed to fetch notes"}), response.status_code


@app.route("/api/note/<int:note_id>")
def get_note(note_id):
    """Получает HTML-версию заметки"""
    headers = {"Authorization": API_TOKEN}
    response = requests.get(f"{SERVER_URL}/notes/{note_id}", headers=headers)

    if response.status_code == 200:
        note = response.json()
        return jsonify({"id": note["id"], "content": note["content"]})
    else:
        return jsonify({"error": "Failed to fetch note"}), response.status_code


if __name__ == "__main__":
    app.run(debug=True)