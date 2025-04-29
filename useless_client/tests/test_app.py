import unittest
from unittest.mock import patch, MagicMock
from useless_client.app import app

from flask import session

def login_session(client):
    with client.session_transaction() as sess:
        sess["logged_in"] = True

class AppTestCase(unittest.TestCase):
    def setUp(self):
        self.client = app.test_client()
        app.config['TESTING'] = True

    @patch('useless_client.app.requests.get')
    def test_index(self, mock_get):
        login_session(self.client)
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = [
            {"id": 1, "path": "notes/note1.md", "content": "<h1>Title</h1><p>Text</p>", "hash": "hash1"}
        ]
        response = self.client.get("/")
        self.assertEqual(response.status_code, 200)
        self.assertIn(b"<h1>Title</h1>", response.data)

    @patch('useless_client.app.requests.get')
    def test_edit_page(self, mock_get):
        login_session(self.client)
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = {
            "id": 1, "path": "notes/note1.md", "content": "# Title\n\nText", "hash": "hash1"
        }

        response = self.client.get("/edit/note1.md")

        self.assertEqual(response.status_code, 404)

    @patch('useless_client.app.requests.post')
    @patch('useless_client.app.requests.get')
    def test_add_note(self, mock_get, mock_post):
        login_session(self.client)

        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = [
            {"id": 1, "path": "notes/note1.md", "content": "<h1>Title</h1><p>Text</p>", "hash": "hash1"}
        ]
        mock_post.return_value.status_code = 200
        mock_post.return_value.json.return_value = {
            "id": 1, "path": "notes/note1.md", "content": "<h1>Title</h1><p>Text</p>", "hash": "hash1"
        }

        response = self.client.post("/add", data={
            "title": "note1",
            "content": "# Title\n\nText"
        }, follow_redirects=True)

        self.assertEqual(response.status_code, 200)
        self.assertIn(b"note1", response.data)

    @patch('useless_client.app.requests.delete')
    @patch('useless_client.app.requests.get')
    def test_delete_note(self, mock_get, mock_delete):
        login_session(self.client)
        mock_get.return_value.status_code = 200
        mock_get.return_value.json.return_value = []
        mock_delete.return_value.status_code = 404
        response = self.client.get("/delete/note1.md")
        self.assertEqual(response.status_code, 404)
