---
form: src/components/SurveyVars.vue
---

# 調查變數

調查變數是每次您執行範本時，任務對話方塊都會提出的問題。答案會以變數的形式傳給任務：對 Ansible 而言是額外變數，對 Terraform 和
OpenTofu 而言是 `-var`，對腳本而言是參數，或是環境變數。儲存範本對話方塊時，變數會隨範本一起儲存。同一個對話方塊也用來編輯工作流程的調查，
此時不提供 **Secret** 類型。

## 名稱 {#surveyVarDialog-name}

任務取得答案時所用的名稱：額外變數、Terraform 變數、腳本的 `name=value` 參數或環境變數的名稱。前後的空格會被移除。此欄位為必填。

請使用調查中其他變數都沒有使用、且工具可接受的名稱，例如 `app_version`。答案會取代範本變數群組中的同名變數。

## 標題 {#surveyVarDialog-title}

任務對話方塊中欄位的標籤，例如 `Version to deploy`。它也是範本對話方塊中該變數標籤上的文字。此欄位為必填。

## 描述 {#surveyVarDialog-description}

選填。在任務對話方塊的欄位取得焦點時顯示於其下方的提示：要輸入什麼，或一個範例。

## 類型 {#surveyVarDialog-type}

任務對話方塊要求的答案種類：

- **String**：一行文字。這是預設值。
- **Integer**：只能輸入數字；任務對話方塊會拒絕其他任何內容，包括負號。答案仍會以文字的形式傳給任務：請在 Playbook 中轉換它，例如使用
  `| int`。
- **Secret**：任務對話方塊會遮蔽的文字。答案不會隨任務儲存，因此重新執行時會再次詢問，而 Ansible 會從以 vault 加密的檔案中取得它。
  Secret 沒有預設值。
- **Enum**：從您在下方定義的清單中選取一個值。
- **Text**：多行文字。
- **Select**：從您在下方定義的清單中選取多個值。任務會以清單的形式取得它們；在參數、`-var` 或環境變數中，清單會以 JSON 撰寫，例如
  `["a","b"]`。

變更類型會清除預設值，而沒有清單的類型會清除各個值。

## Pass variable as {#surveyVarDialog-target}

答案傳給任務的方式：

- **Extra variable**：工具的一般方式。Ansible 以額外變數取得它，Terraform 和 OpenTofu 以 `-var name=value`
  取得，Bash、PowerShell 或 Python 腳本則以腳本之後的 `name=value` 參數取得。
- **Environment variable**：程序的環境變數 `name=value`，適用於每一種範本。此時變數不會再以另一種方式傳遞。

## 鍵 {#surveyVarDialog-valueName}

當類型為 **Enum** 或 **Select** 時顯示。

任務對話方塊在清單中為此值顯示的文字。每個值都需要一個鍵，而且各個鍵必須不同：否則對話方塊會拒絕儲存變數。

## 值 {#surveyVarDialog-valueValue}

當類型為 **Enum** 或 **Select** 時顯示。

選取此值時任務所取得的內容。它可以與其鍵不同：鍵 `Production` 可以傳遞 `prod`。

## Add Value {#surveyVarDialog-addValue}

當類型為 **Enum** 或 **Select** 時顯示。

在值的清單中新增一個空白列。**Enum** 或 **Select** 至少需要一個值。每一列的垃圾桶圖示可將其移除。

## 預設值 {#surveyVarDialog-defaultValue}

除了 **Secret** 之外的每種類型都會顯示。

任務對話方塊開啟時填入的答案；您可以在該處變更它。對於 **Enum**，請選擇其中一個值；對於 **Select**，可選擇任意數量的值。
將此欄位留空，即以空白答案開始。

只有任務對話方塊會填入預設值。以其他方式啟動的任務（例如透過 API）不會取得預設值。

## 必填 {#surveyVarDialog-required}

在變數有答案之前，任務對話方塊不會啟動任務。與預設值一樣，此檢查只由任務對話方塊進行，而非由伺服器進行。
