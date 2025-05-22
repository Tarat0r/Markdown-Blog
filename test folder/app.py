from flask import Flask, render_template, request, redirect, url_for, session, jsonify
import requests
import os
from dotenv import load_dotenv


load_dotenv()
# Настройки скрипта
API_URL = os.getenv("SERVER_URL")  # URL API сервера
API_TOKEN = os.getenv("API_TOKEN")  # Токен API

app = Flask(__name__)
app.secret_key = "your_secret_key"
TEMP_DIR = "temp"

os.makedirs(TEMP_DIR, exist_ok=True)

headers = {
    "Authorization": API_TOKEN
}



@app.route("/")
def index():
    """Главная страница, загружающая пустой шаблон"""
    return render_template("blog.html")


@app.route("/api/notes")
def get_notes():
    """Возвращает список заметок с сервера"""
    headers = {"Authorization": API_TOKEN}
    response = requests.get(f"{API_URL}/notes", headers=headers)

    if response.status_code == 200:
        notes = response.json()
        return jsonify(notes)
    else:
        return jsonify({"error": "Failed to fetch notes"}), response.status_code


@app.route("/api/note/<int:note_id>")
def get_note(note_id):
    """Получает HTML-версию заметки"""
    headers = {"Authorization": API_TOKEN}
    response = requests.get(f"{API_URL}/notes/{note_id}", headers=headers)

    if response.status_code == 200:
        note = response.json()
        return jsonify({"id": note["id"], "content": note["content"]})
    else:
        return jsonify({"error": "Failed to fetch note"}), response.status_code


@app.route("/login", methods=["GET", "POST"])
def login():
    if request.method == "POST":
        username = request.form.get("username")
        password = request.form.get("password")

        if username == "admin" and password == "admin":
            session["logged_in"] = True
            return redirect(url_for("admin"))  # или куда ты хочешь
        else:
            error = "Неверный логин или пароль"
            return render_template("login.html", error=error)

    return render_template("login.html")



@app.route('/admin')
def admin():
    if not session.get("logged_in"):
        return redirect(url_for("login"))
    response = requests.get(f"{API_URL}/notes", headers=headers)
    notes = response.json() if response.ok else []
    return render_template('index.html', notes=notes)


@app.route('/add', methods=['GET', 'POST'])
def add_note():
    if not session.get("logged_in"):
        return redirect(url_for("login"))
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

        responce = requests.post(f"{API_URL}/notes", headers=headers, files=files)
        print(responce.content)
        return redirect(url_for('admin'))

    return render_template('add_note.html')


@app.route('/edit/<int:note_id>', methods=['GET', 'POST'])
def edit_note(note_id):
    if not session.get("logged_in"):
        return redirect(url_for("login"))
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

        responce = requests.put(f"{API_URL}/notes/{note_id}", headers=headers, files=files)
        print(responce.text)
        return redirect(url_for('admin'))

    edit_headers = headers.copy()
    edit_headers["content_md"] = "true"
    response = requests.get(f"{API_URL}/notes/{note_id}", headers=edit_headers)
    note = response.json()
    print(note)
    return render_template('edit_note.html', note=note)

@app.route('/view/<int:note_id>')
def view_note(note_id):
    if not session.get("logged_in"):
        return redirect(url_for("login"))
    view_headers = headers.copy()
    response = requests.get(f"{API_URL}/notes/{note_id}", headers=view_headers)

    if response.status_code != 200:
        return f"Ошибка получения заметки: {response.text}", 400

    note = response.json()
    content = note.get("content", "")  # HTML-контент
    filename = os.path.basename(note["path"])
    title = filename.replace(".md", "").replace("_", " ")

    return render_template('view_note.html', title=title, content=content)


@app.route('/delete/<int:note_id>', methods=['POST'])
def delete_note(note_id):
    if not session.get("logged_in"):
        return redirect(url_for("login"))
    requests.delete(f"{API_URL}/notes/{note_id}", headers=headers)
    return redirect(url_for('admin'))

@app.route("/logout")
def logout():
    session.pop("logged_in", None)
    return redirect(url_for("login"))

if __name__ == '__main__':
    app.run( host="0.0.0.0", port=5050, debug=True)
