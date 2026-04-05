from __future__ import annotations


def evaluate_answer(
    answer_text: str,
    correct_keywords: list[str],
    question_points: int,
) -> tuple[bool, int]:
    """Проверяет, встречается ли хотя бы одно ключевое слово в ответе (без учёта регистра).

    Возвращает (успех, начисленные баллы): при успехе — question_points, иначе 0.
    """
    lowered = answer_text.lower()
    for kw in correct_keywords:
        if kw.lower() in lowered:
            return True, question_points
    return False, 0
