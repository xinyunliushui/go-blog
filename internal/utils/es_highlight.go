package utils

/** 提取高亮文本，如果没有高亮则返回兜底文本
 * @param highlights []string ES返回的高亮数组（例如 []string{"...<mark>Go</mark>语言..."}）
 * @param fallback string 兜底文本（例如数据库里的原始摘要）
 * @return string 高亮文本
 */
func HighlightOrFallback(highlights []string, fallback string) string {
	// 1. 如果高亮数组存在且至少有一个元素，返回第一个高亮片段
	if len(highlights) > 0 {
		return highlights[0]
	}
	// 2. 否则，返回原始文本作为兜底
	return fallback
}
