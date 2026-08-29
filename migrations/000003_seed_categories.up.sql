INSERT INTO categories (name) VALUES
    ('Электроника'),
    ('Одежда'),
    ('Книги'),
    ('Дом и сад'),
    ('Спорт'),
    ('Красота')
ON CONFLICT (name) DO NOTHING;
