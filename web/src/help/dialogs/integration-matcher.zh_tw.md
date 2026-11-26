---
form: src/components/IntegrationMatcherForm.vue
---

# Edit Matcher

比對器是針對傳入請求的條件。比對器決定送往專案別名的請求要啟動哪個整合，因此一個 URL 可以服務多個整合：
只有當整合的所有比對器都成立時，該整合才會執行。只有在整合開啟**使用專案別名**時才會使用比對器；開啟後，該整合只能透過專案別名觸發。

透過專案別名找到但沒有任何比對器的整合永遠不會執行：請至少新增一個比對器。

## Name {#integrationMatcherForm-name}

比對器在整合頁面上列出時所用的名稱。它不會改變比對器檢查的內容。此欄位為必填。

## Match on {#integrationMatcherForm-matchType}

比對器讀取的請求部分：

- **Header**：某個請求標頭的值。
- **Body**：請求的主體。

## Body Data Format {#integrationMatcherForm-bodyDataType}

當 **Match on** 為 **Body** 時顯示。

主體的讀取方式：

- **JSON**：主體會被解析為 JSON，而 **Key** 是指向其中某個值的路徑。
- **String**：整個主體會被當作文字比較，不使用 **Key**。

## Key {#integrationMatcherForm-key}

要從請求中讀取的內容。此欄位為必填。

- 對於 **Header**：標頭的名稱，例如 `X-GitHub-Event`。大小寫無關緊要。
- 對於 **JSON** 格式的 **Body**：值的路徑，各部分以句點連接；陣列的項目寫成 `.[0]`，例如
  `repository.full_name` 或 `commits.[0].id`。
- 對於 **String** 格式的 **Body**：不使用，但表單需要一個值；請隨意輸入。

## Comparison Method {#integrationMatcherForm-method}

從請求讀取的值與 **Value** 的比較方式：

- **==**：兩者相等。
- **!=**：兩者不同。請求中缺少的標頭或鍵會被視為不同。
- **Contains**：從請求讀取的值包含 **Value**。

## Value {#integrationMatcherForm-value}

用來比較的文字，例如 `push` 或 `refs/heads/main`。比較是完全相符且區分大小寫的。JSON 主體中的數字或布林值會以其文字形式比較：
`42`、`true`。此欄位為必填。
