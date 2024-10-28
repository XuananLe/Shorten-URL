from locust import FastHttpUser, TaskSet, between, task

class UserBehavior(TaskSet):
    @task
    def shorten_url(self):
        self.client.get("/short/XPwhgzM7")

class WebsiteUser(FastHttpUser):  
    tasks = [UserBehavior]
    wait_time = between(1, 3)
