#!/bin/bash
sqlite3 ygl.db "INSERT INTO Filter (name, value) 
VALUES 
    ('BedsMin', 2), 
    ('RentMax', 3400), 
    ('DateMin', '2025-07-01'),  
    ('DateMax', '2025-09-01'),
    ('BathsMin', 1.5);"