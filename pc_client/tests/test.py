import unittest
import os
import json
import hashlib
from unittest.mock import patch, mock_open, MagicMock
from pc_client.main import (
    get_string_hash, load_previous_state, save_current_state,
    scan_directory, replace_links, upload_file, update_file, delete_file, sync_files
)


class TestSyncFunctions(unittest.TestCase):

    def test_get_string_hash(self):
        """Проверяем, что хэш строки считается правильно"""
        content = b"test content"
        expected_hash = hashlib.sha256(content).hexdigest()
        self.assertEqual(get_string_hash(content), expected_hash)

    @patch("pc_client.main.os.path.exists", return_value=True)
    @patch("builtins.open", new_callable=mock_open, read_data='{"test.md": {"mtime": 12345}}')
    def test_load_previous_state(self, mock_file, mock_exists):
        """Проверяем загрузку состояния файлов"""
        state = load_previous_state()
        self.assertEqual(state, {"test.md": {"mtime": 12345}})

    @patch("builtins.open", new_callable=mock_open)
    def test_save_current_state(self, mock_file):
        """Проверяем сохранение состояния файлов"""
        state = {"test.md": {"mtime": 12345}}
        save_current_state(state)
        expected = json.dumps(state, indent=4)
        print(expected)
        mock_file().write.assert_called_once_with(expected)

    @patch("pc_client.main.os.walk")
    @patch("pc_client.main.os.path.getmtime", return_value=12345)
    def test_scan_directory(self, mock_mtime, mock_walk):
        """Проверяем сканирование директории"""
        mock_walk.return_value = [("/notes", [], ["test.md"])]
        with patch.dict(os.environ, {"DIRECTORY": "/notes"}):
            result = scan_directory()
        self.assertEqual(result, {os.path.join("/notes","test.md"): {"mtime": 12345}})

    @patch("pc_client.main.find_file", return_value="/full/path/to/file.md")
    def test_replace_links(self, mock_find_file):
        """Проверяем замену Obsidian-ссылок на полные пути"""
        content = "[[note.md|My Note]] and ![[image.png]]"
        expected_result = "[[/full/path/to/file.md|My Note]] and ![[/full/path/to/file.md]]"
        replaced_content, images = replace_links(content)
        self.assertEqual(replaced_content, expected_result)
        self.assertEqual(images, ["/full/path/to/file.md"])

    @patch("pc_client.main.requests.post")
    def test_upload_file(self, mock_post):
        """Проверяем загрузку файла на сервер"""
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_post.return_value = mock_response
        with patch.dict(os.environ, {"API_TOKEN": "fake_token", "SERVER_URL": "http://fake-server"}):
            upload_file("test.md", "content", [])
        mock_post.assert_called_once()

    @patch("pc_client.main.requests.put")
    def test_update_file(self, mock_put):
        """Проверяем обновление файла на сервере"""
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_put.return_value = mock_response
        with patch.dict(os.environ, {"API_TOKEN": "fake_token", "SERVER_URL": "http://fake-server"}):
            update_file("123", "test.md", "content", [])
        mock_put.assert_called_once()

    @patch("pc_client.main.requests.delete")
    def test_delete_file(self, mock_delete):
        """Проверяем удаление файла с сервера"""
        mock_response = MagicMock()
        mock_response.status_code = 200
        mock_delete.return_value = mock_response
        with patch.dict(os.environ, {"API_TOKEN": "fake_token", "SERVER_URL": "http://fake-server"}):
            delete_file("123", "test.md")
        mock_delete.assert_called_once()

    @patch("pc_client.main.get_uploaded_files", return_value=[])
    @patch("pc_client.main.scan_directory", return_value={})
    @patch("pc_client.main.load_previous_state", return_value={})
    @patch("pc_client.main.save_current_state")
    def test_sync_files(self, mock_save_state, mock_load_state, mock_scan, mock_get_uploaded):
        """Проверяем, что синхронизация вызывается без ошибок"""
        sync_files()
        mock_scan.assert_called_once()
        mock_get_uploaded.assert_called_once()
        mock_save_state.assert_called_once()


if __name__ == "__main__":
    unittest.main()
