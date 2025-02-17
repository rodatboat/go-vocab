import client, time, random, requests, sys


vocabClient = client.Client()

if __name__ == "__main__":
    total_sleep = 0
    sleep_limit = 7200 # 3600 = 1 hour, 7200 = 2 hours
    max_points = 100000
    started = False
    correct_answers = 0
    wrong_answers = 0
    
    while True and total_sleep < sleep_limit and vocabClient.total_points < max_points:
        if wrong_answers <= 0:
            percent_correct = 100
        else:
            percent_correct = (correct_answers/(wrong_answers+correct_answers))*100
        print(f"Correct answers: {correct_answers} | Wrong answers: {wrong_answers} | Errors: {vocabClient.total_errors} | Points: {vocabClient.total_points} | Correct %: {round(percent_correct, 2)}% | List Progress %: {round(vocabClient.list_progress*100, 2)}%", end="\r")
        if started:
            sleep_time = round(random.uniform(1, 5), 3) * 0
            time.sleep(sleep_time)
            total_sleep += sleep_time
        elif vocabClient.r_secret == None:
            vocabClient.start_from_list(vocabClient.sat_lists[random.randint(0, len(vocabClient.sat_lists)-1)])
        started = True
        
        fail_counter = 0
        fetch_success = vocabClient.fetched_question_success()
        if not fetch_success:
            if(fail_counter > 5):
                print("Failed to fetch question 5 times. Exiting...")
                exit()
            
            fail_counter += 1
            time.sleep(1)
            fetch_success = vocabClient.fetched_question_success()
        
        answer = vocabClient.askLLM()
        result = vocabClient.answerQuestion(answer)
        if result:
            correct_answers += 1
        else:
            wrong_answers += 1