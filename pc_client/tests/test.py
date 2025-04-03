import unittest
from unittest.mock import patch
import json
from web_client.main import app

class WebClientTestCase(unittest.TestCase):

    @classmethod
    def setUpClass(cls):
        """Настройка тестового клиента"""
        cls.client = app.test_client()

    def test_index(self):
        """Тест рендера главной страницы"""
        response = self.client.get("/")
        self.assertEqual(response.status_code, 200)
        self.assertIn(b"<html", response.data)  # Проверяем, что возвращается HTML

    @patch("requests.get")
    def test_get_notes_success(self, mock_get):
        """Тест успешного получения списка заметок"""
        mock_data = [{"id": 1, "title": "Test Note"}]
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = mock_data

        response = self.client.get("/api/notes")
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json, mock_data)

    @patch("requests.get")
    def test_get_notes_fail(self, mock_get):
        """Тест ошибки при получении списка заметок"""
        mock_get.return_value.status_code = 500

        response = self.client.get("/api/notes")
        self.assertEqual(response.status_code, 500)
        self.assertEqual(response.json, {"error": "Failed to fetch notes"})

    @patch("requests.get")
    def test_get_note_success(self, mock_get):
        """Тест успешного получения одной заметки"""
        note_id = 1
        mock_note = {"id": note_id, "content": "Test Content"}
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = mock_note

        response = self.client.get(f"/api/note/{note_id}")
        self.assertEqual(response.status_code, 200)
        self.assertEqual(response.json, {"id": note_id, "content": "Test Content"})

    @patch("requests.get")
    def test_get_note_fail(self, mock_get):
        """Тест ошибки при получении заметки"""
        note_id = 1
        mock_get.return_value.status_code = 404

        response = self.client.get(f"/api/note/{note_id}")
        self.assertEqual(response.status_code, 404)
        self.assertEqual(response.json, {"error": "Failed to fetch note"})

if __name__ == "__main__":
    unittest.main()
