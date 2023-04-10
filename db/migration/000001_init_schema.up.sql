CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "phone" varchar UNIQUE NOT NULL,
  "role" varchar NOT NULL,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" Date NOT NULL DEFAULT (now())
);

CREATE TABLE "homestays" (
  "id" bigserial PRIMARY KEY,
  "description" varchar NOT NULL,
  "address" varchar NOT NULL,
  "number_of_bed" int NOT NULL,
  "capacity" int NOT NULL,
  "price" decimal NOT NULL,
  "status" varchar NOT NULL DEFAULT 'available',
  "main_image" varchar NOT NULL,
  "first_image" varchar NOT NULL,
  "second_image" varchar NOT NULL,
  "third_image" varchar NOT NULL
);

CREATE TABLE "promotions" (
  "id" bigserial PRIMARY KEY,
  "title" varchar NOT NULL,
  "description" varchar NOT NULL,
  "discount_percent" float NOT NULL,
  "start_date" Date NOT NULL DEFAULT 'now()',
  "end_date" Date NOT NULL DEFAULT '9999-01-01 00:00:00Z'
);

CREATE TABLE "payments" (
  "id" bigserial PRIMARY KEY,
  "amount" decimal NOT NULL,
  "pay_date" Date NOT NULL,
  "pay_method" varchar NOT NULL DEFAULT 'cash',
  "status" varchar NOT NULL DEFAULT 'unpaid'
);

CREATE TABLE "bookings" (
  "id" bigserial PRIMARY KEY,
  "user_booking" varchar NOT NULL,
  "homestay_booking" bigserial NOT NULL,
  "promotion_id" bigserial NOT NULL,
  "payment_id" bigserial NOT NULL,
  "status" varchar NOT NULL DEFAULT 'validated',
  "booking_date" Date NOT NULL DEFAULT 'now()',
  "checkin_date" Date NOT NULL,
  "checkout_date" Date NOT NULL,
  "number_of_guest" int NOT NULL,
  "service_fee" decimal NOT NULL,
  "tax" decimal NOT NULL
);

CREATE TABLE "feedbacks" (
  "id" bigserial PRIMARY KEY,
  "user_comment" varchar NOT NULL,
  "homestay_commented" bigserial NOT NULL,
  "rating" varchar NOT NULL,
  "commention" varchar NOT NULL,
  "created_at" Date NOT NULL DEFAULT 'now()'
);

CREATE INDEX ON "bookings" ("user_booking");

CREATE INDEX ON "bookings" ("homestay_booking");

CREATE INDEX ON "bookings" ("promotion_id");

CREATE INDEX ON "bookings" ("payment_id");

CREATE UNIQUE INDEX ON "bookings" ("homestay_booking", "payment_id", "checkin_date");

CREATE INDEX ON "feedbacks" ("user_comment");

CREATE INDEX ON "feedbacks" ("homestay_commented");

COMMENT ON COLUMN "bookings"."service_fee" IS 'must be positive';

COMMENT ON COLUMN "bookings"."tax" IS 'must be positive';

ALTER TABLE "bookings" ADD FOREIGN KEY ("user_booking") REFERENCES "users" ("username");

ALTER TABLE "bookings" ADD FOREIGN KEY ("homestay_booking") REFERENCES "homestays" ("id");

ALTER TABLE "bookings" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotions" ("id");

ALTER TABLE "bookings" ADD FOREIGN KEY ("payment_id") REFERENCES "payments" ("id");

ALTER TABLE "feedbacks" ADD FOREIGN KEY ("user_comment") REFERENCES "users" ("username");

ALTER TABLE "feedbacks" ADD FOREIGN KEY ("homestay_commented") REFERENCES "homestays" ("id");
