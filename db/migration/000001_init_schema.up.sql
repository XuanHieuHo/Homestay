CREATE TABLE "users" (
  "username" varchar PRIMARY KEY,
  "hashed_password" varchar NOT NULL,
  "full_name" varchar NOT NULL,
  "email" varchar UNIQUE NOT NULL,
  "phone" varchar UNIQUE NOT NULL,
  "role" varchar NOT NULL,
  "isBooking" boolean NOT NULL DEFAULT false,
  "password_changed_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z',
  "created_at" Date NOT NULL DEFAULT (now()),
  "reset_password_token" varchar NOT NULL DEFAULT 'abc',
  "rspassword_token_expired_at" timestamptz NOT NULL DEFAULT '0001-01-01 00:00:00Z'
);

CREATE TABLE "homestays" (
  "id" bigserial PRIMARY KEY,
  "description" varchar NOT NULL,
  "address" varchar NOT NULL,
  "number_of_bed" int NOT NULL,
  "capacity" int NOT NULL,
  "price" float NOT NULL,
  "status" varchar NOT NULL DEFAULT 'available',
  "main_image" varchar NOT NULL,
  "first_image" varchar NOT NULL,
  "second_image" varchar NOT NULL,
  "third_image" varchar NOT NULL
);

CREATE TABLE "promotions" (
  "id" bigserial PRIMARY KEY,
  "title" varchar UNIQUE NOT NULL,
  "description" varchar NOT NULL,
  "discount_percent" float NOT NULL,
  "start_date" Date NOT NULL DEFAULT 'now()',
  "end_date" Date NOT NULL DEFAULT '9999-01-01 00:00:00Z'
);

CREATE TABLE "payments" (
  "id" bigserial PRIMARY KEY,
  "booking_id" varchar NOT NULL,
  "amount" float NOT NULL,
  "pay_date" Date NOT NULL,
  "pay_method" varchar NOT NULL DEFAULT 'cash',
  "status" varchar NOT NULL DEFAULT 'unpaid'
);

CREATE TABLE "bookings" (
  "booking_id" varchar PRIMARY KEY,
  "user_booking" varchar NOT NULL,
  "homestay_booking" bigserial NOT NULL,
  "promotion_id" varchar NOT NULL,
  "status" varchar NOT NULL DEFAULT 'validated',
  "booking_date" Date NOT NULL DEFAULT 'now()',
  "checkin_date" Date NOT NULL,
  "checkout_date" Date NOT NULL,
  "number_of_guest" int NOT NULL,
  "service_fee" float NOT NULL,
  "tax" float NOT NULL
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

CREATE INDEX ON "feedbacks" ("user_comment");

CREATE INDEX ON "feedbacks" ("homestay_commented");

COMMENT ON COLUMN "bookings"."service_fee" IS 'must be positive';

COMMENT ON COLUMN "bookings"."tax" IS 'must be positive';

ALTER TABLE "payments" ADD FOREIGN KEY ("booking_id") REFERENCES "bookings" ("booking_id");

ALTER TABLE "bookings" ADD FOREIGN KEY ("user_booking") REFERENCES "users" ("username");

ALTER TABLE "bookings" ADD FOREIGN KEY ("homestay_booking") REFERENCES "homestays" ("id");

ALTER TABLE "bookings" ADD FOREIGN KEY ("promotion_id") REFERENCES "promotions" ("title");

ALTER TABLE "feedbacks" ADD FOREIGN KEY ("user_comment") REFERENCES "users" ("username");

ALTER TABLE "feedbacks" ADD FOREIGN KEY ("homestay_commented") REFERENCES "homestays" ("id");
