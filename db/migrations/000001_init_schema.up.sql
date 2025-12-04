
CREATE TABLE "user" (
  "id" SERIAL PRIMARY KEY,
  "user_name" VARCHAR (100) UNIQUE NOT NULL,
  "password_hash" TEXT NOT NULL,
  "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
); 


CREATE TABLE "post" (
  "id" SERIAL PRIMARY KEY,
  "user_id" INT NOT NULL REFERENCES "user"("id") ON DELETE CASCADE, 
  "title" VARCHAR(255) NOT NULL,
  "content" TEXT NOT NULL,
  "created_at" TIMESTAMP NOT NULL DEFAULT NOW()
);