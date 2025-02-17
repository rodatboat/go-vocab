CREATE TABLE IF NOT EXISTS question (
    id SERIAL PRIMARY KEY,
    question_type VARCHAR(255) NOT NULL,
    question TEXT NOT NULL,
    question_context TEXT,
    question_code TEXT NOT NULL,
    question_html TEXT NOT NULL,
    answer TEXT NOT NULL,
    answer_data_key VARCHAR(255) NOT NULL,
    difficulty NUMERIC,
    choices TEXT,
    correct BOOLEAN DEFAULT FALSE,
    target_word VARCHAR(255),

    UNIQUE (question_type, question_context, question)
);
