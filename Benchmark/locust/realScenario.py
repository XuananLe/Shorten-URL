import random
from locust import HttpUser, task, between
from urllib3 import PoolManager
import os
import sys

class ShortenUrlUser(HttpUser):
    wait_time = between(10, 20)  
    responded_shortened_urls = []

    @task(1)
    def create_url(self):
        """
        Task to create a shortened URL.
        """
        response = self.client.post(
            "/create",
            params={  
                "url": "https://kubernetes.io/docs/concepts/overview/components/",
                "userId": "1c8be2ab-694d-40a1-acda-6d2ff09e8b76"
            }
        )
        if response.status_code == 201:
            short_url = response.json().get("shortUrl", None)
            if short_url:
                print(f"Shortened URL: {short_url}")
                self.responded_shortened_urls.append(short_url)
        else:
            print(f"Failed to create shortened URL: {response.text}")
        self.wait()

    @task(4)
    def access_url(self):
        if self.responded_shortened_urls:
            short_url = self.responded_shortened_urls[random.randint(0, len(self.responded_shortened_urls) - 1)]
            response = self.client.get(f"/short/{short_url}")
            if response.status_code == 200:
                original_url = response.json().get("originalUrl", None)
                assert original_url == "https://kubernetes.io/docs/concepts/overview/components/", "Original URL mismatch"
                print(f"Accessed URL successfully: {original_url}")
            else:
                print(f"Failed to access shortened URL: {response.status_code}, {short_url}")
        else:
            sys.exit(1);
            os._exit(1);
            print("No shortened URLs to access.")
