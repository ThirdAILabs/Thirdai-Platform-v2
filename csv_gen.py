import pandas as pd
import faker
import random

# Initialize Faker for fake data generation
fake = faker.Faker()

# Define the number of rows and columns
num_rows = 1000
num_columns = 20

# Define movie genres
genres = ["Action", "Comedy", "Drama", "Horror", "Sci-Fi", "Thriller", "Romance", "Adventure", "Fantasy", "Animation"]

# Generate random movie data
data = []
for _ in range(num_rows):
    title = fake.catch_phrase()  # Movie title
    genre = random.choice(genres)  # Genre
    release_year = random.randint(1980, 2023)  # Release year
    director = fake.name()  # Director
    runtime = random.randint(60, 180)  # Runtime in minutes
    rating = round(random.uniform(1.0, 10.0), 1)  # Rating (1.0 to 10.0)
    budget = random.randint(1000000, 200000000)  # Budget in USD
    box_office = random.randint(1000000, 1000000000)  # Box office revenue in USD
    language = random.choice(["English", "Spanish", "French", "German", "Chinese", "Japanese"])  # Language
    country = fake.country()  # Country of origin
    is_sequel = random.choice([True, False])  # Is it a sequel?
    is_remake = random.choice([True, False])  # Is it a remake?
    production_company = fake.company()  # Production company
    lead_actor = fake.name()  # Lead actor
    lead_actress = fake.name()  # Lead actress
    release_date = fake.date_between(start_date="-30y", end_date="today").strftime("%Y-%m-%d")  # Release date
    imdb_id = f"tt{random.randint(1000000, 9999999)}"  # Fake IMDb ID
    awards = random.choice(["Oscar Winner", "Golden Globe Winner", "None"])  # Awards
    streaming_platform = random.choice(["Netflix", "Amazon Prime", "Hulu", "Disney+", "HBO Max"])  # Streaming platform

    row = [
        title, genre, release_year, director, runtime, rating, budget, box_office, language, country,
        is_sequel, is_remake, production_company, lead_actor, lead_actress, release_date, imdb_id, awards, streaming_platform
    ]
    data.append(row)

# Define column names
columns = [
    "Title", "Genre", "Release Year", "Director", "Runtime (min)", "Rating", "Budget (USD)", "Box Office (USD)",
    "Language", "Country", "Is Sequel", "Is Remake", "Production Company", "Lead Actor", "Lead Actress",
    "Release Date", "IMDb ID", "Awards", "Streaming Platform"
]

# Create DataFrame
df = pd.DataFrame(data, columns=columns)

# Save to CSV
df.to_csv("movies_data.csv", index=False)

print("CSV file 'movies_data.csv' generated successfully!")