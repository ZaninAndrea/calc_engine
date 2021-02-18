# Calc engine

This is the engine that parses and executes `.calc` files.

## Features

In a calc file each line is a mathematical expression, that can also be stored and used in other lines. E.g.

```
55 + y
y: sqrt(11+5)+3
```

Supported functions are `sqrt log sin cos tan abs ln round ceil floor`, the software also recognizes the constants `pi e`.
Numbers are expressed with the decimal comma, e.g. `55,2`, you can use the dot to split long numbers, e.g. `1.000.000`, and you can express numbers as percentages, e.g. `56%`.
