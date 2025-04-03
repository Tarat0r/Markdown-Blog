import pytest
import json
from flask import Flask
from web_client.main import app

@pytest.fixture
def client():
    """Создаёт тестовый клиент Flask"""
    app.config["TESTING"] = True
    return app.test_client()

def test_index(client):
    """Тест рендера главной страницы"""
    response = client.get("/")
    assert response.status_code == 200
    assert b"<html" in response.data  # Проверяем, что возвращается HTML

def test_get_notes_success(client, requests_mock):
    """Тест успешного получения списка заметок"""
    mock_data = [{"id": 1, "title": "Test Note"}]
    requests_mock.get("http://mockserver/notes", json=mock_data, status_code=200)

    response = client.get("/api/notes")
    assert response.status_code == 200
    assert response.json == mock_data

def test_get_notes_fail(client, requests_mock):
    """Тест ошибки при получении списка заметок"""
    requests_mock.get("http://mockserver/notes", status_code=500)

    response = client.get("/api/notes")
    assert response.status_code == 500
    assert response.json == {"error": "Failed to fetch notes"}

def test_get_note_success(client, requests_mock):
    """Тест успешного получения одной заметки"""
    note_id = 1
    mock_note = {"id": note_id, "content": "Test Content"}
    requests_mock.get(f"http://mockserver/notes/{note_id}", json=mock_note, status_code=200)

    response = client.get(f"/api/note/{note_id}")
    assert response.status_code == 200
    assert response.json == {"id": note_id, "content": "Test Content"}

def test_get_note_fail(client, requests_mock):
    """Тест ошибки при получении заметки"""
    note_id = 1
    requests_mock.get(f"http://mockserver/notes/{note_id}", status_code=404)

    response = client.get(f"/api/note/{note_id}")
    assert response.status_code == 404
    assert response.json == {"error": "Failed to fetch note"}
