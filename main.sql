CREATE TABLE event (
  id INT GENERATED ALWAYS AS IDENTITY PRIMARY KEY,
  title VARCHAR(128) NOT NULL,
  description VARCHAR(255) NOT NULL,
  start_date TIMESTAMP NOT NULL,
  end_date TIMESTAMP NOT NULL,
  created_at TIMESTAMP DEFAULT NOW()
)
  
INSERT INTO event (title, description, start_date, end_date)
VALUES
('Breakfast', '', '2025-08-31 9:00:00', '2025-08-31 9:30:00'),
('Lunch', '', '2025-08-31 12:30:00', '2025-08-31 13:30:00'),
('Dinner', '', '2025-08-31 19:30:00', '2025-08-31 20:30:00')
