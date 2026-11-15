# Demo task

A tiny, safe task to prove the loop end to end (a few cents on opus).

1. In the working directory, create `slugify.py`: a function `slugify(s)` that lowercases, trims, and replaces every
   run of non-alphanumeric characters with a single hyphen, with no leading or trailing hyphen.
2. Create `test_slugify.py` with at least six cases, including empty string, leading/trailing punctuation, unicode
   letters, and a long run of symbols. Run it with `python3 -m pytest -q` (or `python3 test_slugify.py` if pytest is
   absent) and make it pass.
3. Copy both files into the output directory and reply `FINAL:` with the passing test count.
