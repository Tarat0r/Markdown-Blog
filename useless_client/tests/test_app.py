import unittest
import io
from useless_client.app import app
from unittest.mock import patch
import json

class AppTestCase(unittest.TestCase):

    def setUp(self):
        self.client = app.test_client()
        self.api_token = "a" * 64  # Пример валидного токена

    @patch("useless_client.app.requests.get")
    def test_index_page(self, mock_get):
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = [
            {"id": 1, "path": "notes/note1.md", "hash": "hash1"},
            {"id": 2, "path": "notes/note2.md", "hash": "hash2"}
        ]

        response = self.client.get("/")
        self.assertEqual(response.status_code, 200)
        self.assertIn(b"notes/note1.md", response.data)
        self.assertIn(b"notes/note2.md", response.data)

    @patch("useless_client.app.requests.get")
    def test_edit_page(self, mock_get):
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = {
            "id": 1,
            "path": "notes/note1.md",
            "content": "<h1>Title</h1><p>Text</p>",
            "hash": "hash1"
        }

        response = self.client.get("/edit/1")
        self.assertEqual(response.status_code, 200)
        self.assertIn(b"Title", response.data)

    @patch("useless_client.app.requests.post")
    def test_add_note(self, mock_post):
        mock_post.return_value.status_code = 200
        mock_post.return_value.json.return_value = {"message": "Upload successful"}

        data = {
            "title": "Test Note",
            "content": "# Test Markdown\nSome content"
        }

        response = self.client.post(
            "/add",
            data=data,
            follow_redirects=True
        )

        self.assertEqual(response.status_code, 200)

    @patch("useless_client.app.requests.delete")
    def test_delete_note(self, mock_delete):
        mock_delete.return_value.status_code = 200
        mock_delete.return_value.json.return_value = {"path": "notes/test.md"}

        response = self.client.post("/delete/1", follow_redirects=True)
        self.assertEqual(response.status_code, 200)

if __name__ == "__main__":
    unittest.main()
