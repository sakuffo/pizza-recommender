CREATE TABLE IF NOT EXISTS pizzas (
  id   INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT    NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS ingredients (
  id   INTEGER PRIMARY KEY AUTOINCREMENT,
  name TEXT    NOT NULL UNIQUE
);

CREATE TABLE IF NOT EXISTS pizza_ingredients (
  pizza_id      INTEGER NOT NULL,
  ingredient_id INTEGER NOT NULL,
  FOREIGN KEY (pizza_id)      REFERENCES pizzas(id) ON DELETE CASCADE,
  FOREIGN KEY (ingredient_id) REFERENCES ingredients(id) ON DELETE CASCADE,
  UNIQUE(pizza_id, ingredient_id)
);