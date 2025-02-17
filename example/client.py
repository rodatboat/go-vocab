import json, base64, random, os
from bs4 import BeautifulSoup as soup
from curl_cffi import requests

class Client:
    next_question_endpoint = "https://www.vocabulary.com/challenge/nextquestion.json"
    save_answer_endpoint = "https://www.vocabulary.com/challenge/saveanswer.json"
    start_endpoint = "https://www.vocabulary.com/challenge/start.json"
    llm_endpoint = "http://localhost:11434/api/generate"

    important_cookies = ["AWSALB", "AWSALBCORS", "JSESSIONID", "guid"]
    headers = {
    'User-Agent': 'Mozilla/5.0 (Windows NT 10.0; Win64; x64) AppleWebKit/537.36 (KHTML, like Gecko) Chrome/130.0.0.0 Safari/537.36 OPR/115.0.0.0',
    'Origin': 'https://www.vocabulary.com',
    'X-Requested-With': 'XMLHttpRequest',
    'Content-Type': 'application/x-www-form-urlencoded; charset=UTF-8',
    'Cookie': 'AWSALB=123; guid=123; JSESSIONID=123'
    }
    sat_lists = [8340291, 8995949, 9048293, 9336685, 148703, 151274, 148713, 148732, 148845, 149637, 149640, 149642, 149643, 151263, 151274, 151399, 151404, 151465, 151466, 156619, 156622, 158007, 158769, 158781, 158782, 161539]    

    r_secret = ""
    current_question = {}
    question_type = ""
    # D = Multiple choice (<word> means: <definition>)
    # I = Image multiple choice (<image>)
    # S = Multiple choice (<word> has same or similar meaning (synonym): <word>)
    # L = Multiple choice (<example quote> in this sentence <word> means: <definition>)
    # F = Multiple choice (Complete the sentence with <word>: <word>)

    # RESULTS
    total_points = 0
    total_errors = 0
    list_progress = 0

    def __init__(self):
        self.session = requests.Session()
        self.loadLocalProgress()
        
    def start_from_list(self, listId):
        # print("Initializing vocabulary practice from list...")
        self.listId = listId
        payload = {
            "v": 3,
            "activitytype": "p",  # c = challenge, p = practice
            "wordlistid": self.listId # wordlist
        }
        
        if self.r_secret != "":
            print("Continuing from previous session...")
            # print("Secret:", self.r_secret)
            payload["secret"] = self.r_secret
        
        cookies = {}
        for cookie in self.headers['Cookie'].split('; '):
            key, value = cookie.split('=')
            cookies[key] = value
        data = self.session.post(self.start_endpoint,
                          json=payload, headers=self.headers, cookies=cookies)

        successful_start = data.status_code == 200
        if not successful_start:
            self.total_errors += 1
            # print("Failed to start practice session from list.")
            # print("ERROR:", data.status_code)
            try:
                # print("ERROR MESSAGE:", json.loads(data.text))
                pass
            except:
                raise Exception("ERROR")
        data_json = json.loads(data.text)
        self.r_secret = data_json["secret"]
        self.list_progress = 0
        
        if int(data_json["pdata"]["points"]) < 100000:
            print("Not logged in...")
            exit()

        self.question_type = data_json["qtype"]
        b64_question = data_json["code"]
        question_html = base64.b64decode(b64_question).decode('utf-8')

        print("Question Type:", self.question_type)
        # print("Question (Encoded): ", b64_question)
        
        self.current_question = self.parseQuestion(question_html)
        self.saveLocalProgress()
        data_cookies = data.cookies.get_dict()
        self.updateCookies(data_cookies)
        # print("Question (Decoded): ", self.current_question)

    def parseQuestion(self, htmlData):
        # print("Parsing question...")
        htmlData = soup(htmlData, 'html.parser')

        # Gets the question context, some questions have additional information like a quote or example.
        context = htmlData.find("div", {"class": "questionContent"})
        if context != None:
            context = context.find_all("div", {"class": "sentence"}) if context != None else None
            if len(context) > 0:
                if len(context) == 1:
                    context = self.clean_string(context[0].text)
                else:
                    newContext = []
                    for i in context:
                        newContext.append(self.clean_string(i.text))
                    context = newContext
                    
        if self.question_type == "T":
            correct_answer = htmlData.find("div", {"class": "complete"}).find("strong").text
            return {
                "context":context,
                "question":"",
                "choices":[{
                    "answer": correct_answer,
                    "code": correct_answer
                }],
                "done":False
            }

        # Gets the question itself
        try:
            question = htmlData.find("div", {"class": "instructions"}).text
            question = self.clean_string(question)
        except:
            question = ""
            pass

        # Gets the answer choices
        choices = htmlData.find("div", {"class": "choices"}).find_all("a")

        if self.question_type == "I":
            choices = [{"answer":c["style"], "code":c["data-nonce"]} for c in choices]
        else:
            choices = [{"answer":self.clean_string(c.text), "code":c["data-nonce"]} for c in choices]

        result = {
            "context":context,
            "question":question,
            "choices":choices,
            "done":False
        }

        self.current_question = result
        return result
    
    def answerQuestion(self, answer):
        # print("Answering question...")
        correct_answer = "abc123"
        try:
            correct_answer = answer["answer"]["code"]
        except:
            pass
        
        payload = {
            "secret": self.r_secret,
            "v": 3,
            "rt": int(round(random.uniform(3, 7), 3) * 1000),
            "a": correct_answer
        }

        data = self.session.post(self.save_answer_endpoint, data=payload, headers=self.headers)

        successful_start = data.status_code == 200
        if not successful_start:
            self.total_errors += 1
            print(data.status_code, data.text)
            # print("Failed to answer question.")
            raise Exception("ERROR")
        
        data_json = json.loads(data.text)
        self.r_secret = data_json["secret"]
        
        if int(data_json["pdata"]["points"]) < 100000:
            print("Not logged in...")
            exit()

        # Currently does nothing, but using this dict we can update the current cookies
        data_cookies = data.cookies.get_dict()
        self.updateCookies(data_cookies)

        answer_correct = data_json["answer"]["correct"]
        try:
            self.list_progress = float(data_json["game"]["progress"])
        except:
            self.start_from_list(self.sat_lists[random.randint(0, len(self.sat_lists)-1)])
            pass
        
        if not answer_correct:
            # print("WRONG ANSWER!")
            self.saveLocalProgress()
            return False
        else:
            # print("RIGHT ANSWER!")
            self.total_points += int(data_json["answer"]["points"])
            # print(self.total_points)
            self.saveLocalProgress()
            
            with open("data.txt", "a", encoding="utf-8") as datafile:
                line = '{}|{}|{}|{}\n'.format(self.current_question["context"], self.current_question["question"], self.current_question["choices"], answer)
                datafile.write(line)
            return True
        
    def askLLM(self):
        # print("Asking LLM...")
        
        payload = {
            "model": "llama3.1:8b-instruct-q5_K_S",
            "system":"You're a a vocabulary teacher, that answers my questions about vocabulary. You not modify the question, and will keep the question and answers as received.",
            "prompt": json.dumps(self.current_question),
            "format": {
                "type": "object",
                "properties": {
                "question": {
                    "type": "string"
                },
                "answer": {
                        "type": "object",
                        "properties": {
                            "answer": {
                    "type": "string"
                },
                            "code": {
                    "type": "string"
                }
                        }
                    },
                "choices": {
                    "type": "array",
                    "items": {
                        "type": "object",
                        "properties": {
                            "answer": {
                    "type": "string"
                },
                            "code": {
                    "type": "string"
                }
                        }
                    }
                }
                },
                "required": [
                "question",
                "answer",
                "choices"
                ]
            },
            "stream": False
        }
        
        data = self.session.post(self.llm_endpoint, json=payload)
        
        # print(json.dumps(payload))
        # print(data.text)
        # print(data.status_code)
        
        data_json = json.loads(data.text)
        
        
        answer_json = json.loads(data_json["response"])
        return answer_json
        
    def next_question(self):
        # print("Getting next question...")
        payload = {
            "secret": self.r_secret,
            "v": 3,
        }
        cookies = {}
        for cookie in self.headers['Cookie'].split('; '):
            key, value = cookie.split('=')
            cookies[key] = value
        data = self.session.post(self.next_question_endpoint, json=payload,
                                 headers=self.headers, cookies=cookies)
        
        successful_start = data.status_code == 200
        if not successful_start:
            self.total_errors += 1
            # print("Failed to get next question.")
            # print("ERROR:", data.status_code)
            raise Exception("ERROR")

        data_json = json.loads(data.text)
        self.r_secret = data_json["secret"]
        
        if int(data_json["pdata"]["points"]) < 100000:
            print("Not logged in...")
            exit()

        if "game" in data_json and float(data_json["game"]["progress"]) == 1.0:
            self.list_progress = 1
            self.start_from_list(self.sat_lists[random.randint(0, len(self.sat_lists)-1)])
            return
        else:
            self.question_type = data_json["qtype"]

            # Currently does nothing, but using this dict we can update the current cookies
            data_cookies = data.cookies.get_dict()
            self.updateCookies(data_cookies)

            b64_question = data_json["code"]
            question_html = base64.b64decode(b64_question).decode('utf-8')

            # print("Question Type:", self.question_type)
            # print("Question (Encoded): ", b64_question)
            
            self.current_question = self.parseQuestion(question_html)
            # print("Question (Decoded): ", self.current_question)
            self.list_progress = float(data_json["game"]["progress"])

    def clean_string(self, string):
        if string is not None and isinstance(string, str):
            return string.strip().replace('\r', '').replace('\n', '').replace('\t', ' ')
        else:
            return string
        
    def getCurrentQuestion(self):
        return self.current_question
    
    def saveLocalProgress(self):
        # print("Updating local progress...")
        save = {
            "current_question": self.current_question,
            "points": self.total_points,
            "question_type": self.question_type,
            "r_secret": self.r_secret,
            "cookies": self.headers["Cookie"]
        }
        
        with open("progress.json", "w") as outfile:
            json.dump(save, outfile)
            
    def loadLocalProgress(self):
        if os.path.exists("progress.json"):
            with open("progress.json", "r") as infile:
                data = json.load(infile)
                self.current_question = data["current_question"]
                self.total_points = data["points"]
                self.question_type = data["question_type"]
                self.r_secret = data["r_secret"]
                self.cookies = self.headers["Cookie"]
            
    def fetched_question_success(self):
        try:
            if self.list_progress == 1:
                self.start_from_list(self.sat_lists[random.randint(0, len(self.sat_lists)-1)])
            else:
                self.next_question()
            return True
        except:
            return False
        
    def updateCookies(self, diction):
        newCookies = ""
        for key, value in diction.items():
            if key in self.important_cookies:
                newCookies += f"{key}={value}; "
        self.headers['Cookie'] = newCookies.strip()[:-1]