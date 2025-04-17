from flask import Flask, render_template, request, redirect, url_for
import requests
import os
from dotenv import load_dotenv


load_dotenv()
# Настройки скрипта
API_URL = os.getenv("SERVER_URL", "http://localhost:8080")+"/notes"  # URL API сервера
API_TOKEN = os.getenv("API_TOKEN")  # Токен API

app = Flask(__name__)
TEMP_DIR = "temp"

os.makedirs(TEMP_DIR, exist_ok=True)

headers = {
    "Authorization": API_TOKEN
}


@app.route('/')
def index():
    response = requests.get(API_URL, headers=headers)
    notes = response.json() if response.ok else []
    return render_template('index.html', notes=notes)


@app.route('/add', methods=['GET', 'POST'])
def add_note():
    if request.method == 'POST':
        title = request.form['title']
        content = request.form['content']
        filename = f"{title.replace(' ', '_')}.md"
        filepath = os.path.join(TEMP_DIR, filename)

        # сохранить markdown временно
        with open(filepath, 'w') as f:
            f.write(content)

        # отправить multipart/form-data
        metadata = {
            "path": f"notes/{filename}",
            "images": []
        }

        files = {
            "metadata": (None, str(metadata).replace("'", '"'), 'application/json'),
            "markdown": open(filepath, 'rb')
        }

        responce = requests.post(API_URL, headers=headers, files=files)
        print(responce.content)
        return redirect(url_for('index'))

    return render_template('add_note.html')


@app.route('/edit/<int:note_id>', methods=['GET', 'POST'])
def edit_note(note_id):
    if request.method == 'POST':
        title = request.form['title']
        content = request.form['content']
        filename = f"{title.replace(' ', '_')}.md"
        filepath = os.path.join(TEMP_DIR, filename)

        with open(filepath, 'w') as f:
            f.write(content)

        metadata = {
            "path": f"notes/{filename}",
            "images": []
        }

        files = {
            "metadata": (None, str(metadata).replace("'", '"'), 'application/json'),
            "markdown": open(filepath, 'rb')
        }

        responce = requests.put(f"{API_URL}/{note_id}", headers=headers, files=files)
        print(responce.text)
        return redirect(url_for('index'))

    edit_headers = headers.copy()
    edit_headers["content_md"] = "true"
    response = requests.get(f"{API_URL}/{note_id}", headers=edit_headers)
    note = response.json()
    print(note)
    return render_template('edit_note.html', note=note)

@app.route('/view/<int:note_id>')
def view_note(note_id):
    view_headers = headers.copy()
    response = requests.get(f"{API_URL}/{note_id}", headers=view_headers)

    if response.status_code != 200:
        return f"Ошибка получения заметки: {response.text}", 400

    note = response.json()
    content = note.get("content", "")  # HTML-контент
    filename = os.path.basename(note["path"])
    title = filename.replace(".md", "").replace("_", " ")

    return render_template('view_note.html', title=title, content=content)
@app.route('/delete/<int:note_id>', methods=['POST'])
def delete_note(note_id):
    requests.delete(f"{API_URL}/{note_id}", headers=headers)
    return redirect(url_for('index'))


if __name__ == '__main__':
    app.run(port=5050, debug=True)
