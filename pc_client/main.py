import os
import hashlib
import json
import re
import requests

# Настройки скрипта
DIRECTORY = "C:/Users/Kevin/Downloads/Markdown-Blog-main/Markdown-Blog/test"  # Директория с Markdown-файлами
SERVER_URL = "http://localhost:8080"  # URL API сервера
STATE_FILE = "file_state.json"  # Файл для хранения состояния
API_TOKEN = "xFRFXi9fn3xwsOS8di8abUfouKJs6HqrrdejCkW1cvTm9XiCo6B49YeZL9HvHvdK"  # Токен API


def get_string_hash(content):
    """Вычисляет хэш содержимого файла для определения изменений."""
    hasher = hashlib.sha256()
    hasher.update(content.encode("utf-8"))
    return hasher.hexdigest()


def load_previous_state():
    """Загружает сохранённое состояние файлов."""
    if os.path.exists(STATE_FILE):
        with open(STATE_FILE, "r") as f:
            return json.load(f)
    return {}


def save_current_state(state):
    """Сохраняет текущее состояние файлов."""
    with open(STATE_FILE, "w") as f:
        json.dump(state, f, indent=4)


def scan_directory():
    """Сканирует директорию и собирает информацию о файлах."""
    file_data = {}
    for root, _, files in os.walk(DIRECTORY):
        for file in files:
            if file.endswith(".md"):
                path = os.path.join(root, file)
                file_data[path] = {"mtime": os.path.getmtime(path)}
    return file_data


def get_uploaded_files():
    """Запрашивает у сервера список уже загруженных файлов."""
    headers = {"Authorization": API_TOKEN}
    response = requests.get(f"{SERVER_URL}/notes", headers=headers)
    print(response.text)

    return response.json() if response.status_code == 200 else []


def replace_links(content, file_dir):
    """Заменяет Obsidian-ссылки на полные пути."""

    def note_replacer(match):
        note_name, text = match.groups()
        full_path = os.path.abspath(os.path.join(file_dir, f"{note_name}.md"))
        return f"[[{full_path}|{text}]]"

    def image_replacer(match):
        img_name = match.group(1)
        full_path = os.path.abspath(os.path.join(file_dir, img_name))
        return f"![[{full_path}]]"

    content = re.sub(r"\[\[([^|\]]+)\|([^\]]+)\]\]", note_replacer, content)
    content = re.sub(r"!\[\[([^\]]+)\]\]", image_replacer, content)

    return content


def update_file(note_id, path, content):
    """Обновляет существующую заметку на сервере."""
    headers = {"Authorization": API_TOKEN, "Content-Type": "application/json"}
    data = json.dumps({"path": path, "content": content})
    response = requests.put(f"{SERVER_URL}/notes/{note_id}", headers=headers, data=data)
    if response.status_code == 200:
        print(f"Updated: {path}")
    else:
        print(f"Failed to update {path}: ({response.status_code}) {response.text}")


def delete_file(note_id, path):
    """Удаляет файл с сервера."""
    headers = {"Authorization": API_TOKEN}
    response = requests.delete(f"{SERVER_URL}/notes/{note_id}", headers=headers)
    if response.status_code == 200:
        print(f"Deleted: {path}")
    else:
        print(f"Failed to delete {path}: ({response.status_code}) {response.text}")


def upload_file(path, content):
    """Загружает новый файл на сервер."""
    metadata = json.dumps({"path": path, "images": []})
    headers = {"Authorization": API_TOKEN}
    files = {
        "metadata": (None, metadata, "application/json"),
        "markdown": (os.path.basename(path), content, "text/plain")
    }
    response = requests.post(f"{SERVER_URL}/notes", headers=headers, files=files)
    if response.status_code == 200:
        print(f"Uploaded: {path} {response.text}")
    else:
        print(f"Failed to upload {path}: ({response.status_code}) {response.text}")


def sync_files():
    """Синхронизирует файлы с сервером."""
    previous_state = {} #load_previous_state()
    current_state = scan_directory()
    uploaded_files = {}#{note["path"]: (note["id"], note["hash"]) for note in get_uploaded_files()}

    for path, data in current_state.items():
        filename = os.path.basename(path)
        with open(path, "r", encoding="utf-8") as f:
            content = f.read()

        modified_content = replace_links(content, os.path.dirname(path))
        content_hash = get_string_hash(modified_content)

        if path in uploaded_files:
            note_id, server_hash = uploaded_files[path]
            if server_hash != content_hash:
                print(f"Updating file: {filename}")
                #update_file(note_id, path, modified_content)
        else:
            print(f"Uploading new file: {filename}")
            upload_file(path, modified_content)

    for path in previous_state:
        if path not in current_state and path in uploaded_files:
            print(f"Deleting file: {os.path.basename(path)}")
            delete_file(uploaded_files[path][0], path)

    save_current_state(current_state)


if __name__ == "__main__":
    sync_files()
