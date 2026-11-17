---
form: src/components/IntegrationExtractValueForm.vue
---

# Edit Extract Value

擷取值會取出傳入請求的一部分，例如 push 事件的分支，並將其傳給整合所啟動的任務。整合的每個值都會從每個啟動任務的請求中取得。

## Name {#integrationValueForm-name}

此值在整合頁面上列出時所用的名稱。它不會改變取出的內容或其去向。此欄位為必填。

## Source of the Value {#integrationValueForm-valueSource}

取得此值的請求部分：

- **Header**：某個請求標頭的值。
- **Body**：請求的主體。

## Data Type of Body {#integrationValueForm-bodyDataType}

當 **Source of the Value** 為 **Body** 時顯示。

主體的讀取方式：

- **JSON**：主體會被解析為 JSON，而 **Key** 是指向其中某個值的路徑。此值會保留其 JSON 類型：數字、清單或物件會原樣傳給任務。
- **String**：整個主體會被當作文字取出，不使用 **Key**。

## Key {#integrationValueForm-key}

要從請求中取出的內容。此欄位為必填。

- 對於 **Header**：標頭的名稱，例如 `X-GitHub-Event`。大小寫無關緊要。請求中缺少的標頭會產生空值。
- 對於 **JSON** 格式的 **Body**：值的路徑，各部分以句點連接；陣列的項目寫成 `.[0]`，例如
  `repository.full_name` 或 `commits.[0].id`。主體中不存在的路徑不會產生任何值，任務會保留整合中設定的值。
- 對於 **String** 格式的 **Body**：不使用，但表單需要一個值；請隨意輸入。

## Variable Usage {#integrationValueForm-variableType}

此值在範本的任務中的去向：

- **Variables**：任務的額外變數，傳遞方式與調查變數的答案相同。它會取代同名的答案或變數群組中的同名變數。
- **Task Params**：任務的參數，儲存方式與任務對話方塊的選項相同。當範本允許覆寫分支時，名稱 `git_branch`
  會設定任務的分支；開頭的 `refs/heads/` 會被移除。

執行工作流程的整合會忽略此用途：每個值都會成為工作流程中以 **Variable** 命名的調查變數的答案。

## Variable {#integrationValueForm-variable}

傳遞此值時所用的名稱，例如 `branch`，作為 Playbook 以 `{{ branch }}` 讀取的額外變數。此欄位為必填。
