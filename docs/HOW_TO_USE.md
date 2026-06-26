1|# 📖 Coding Challenge Platform - Panduan Penggunaan
2|
3|> Panduan lengkap menggunakan platform coding challenge untuk user, admin, dan developer.
4|
5|---
6|
7|## 📋 Daftar Isi
8|
9|1. [Untuk User (Peserta)](#untuk-user-peserta)
10|2. [Untuk Admin](#untuk-admin)
11|3. [API Reference](#api-reference)
12|4. [Contoh Kode & Submission](#contoh-kode--submission)
13|5. [Keyboard Shortcuts](#keyboard-shortcuts)
14|6. [FAQ](#faq)
15|
16|---
17|
18|## 🧑‍💻 Untuk User (Peserta)
19|
20|### 1. Registrasi & Login
21|
22|#### Registrasi Akun Baru
23|
24|1. Buka `http://localhost:9100` di browser
25|2. Klik tombol **"Register"** di pojok kanan atas
26|3. Isi form registrasi:
27|   - **Username**: Nama unik (3-50 karakter, huruf/angka/_)
28|   - **Email**: Email valid
29|   - **Password**: Minimal 8 karakter (huruf besar, huruf kecil, angka)
30|   - **Full Name**: Nama lengkap
31|4. Klik **"Create Account"**
32|5. Anda akan otomatis login dan diarahkan ke halaman utama
33|
34|#### Login
35|
36|1. Klik **"Login"** di pojok kanan atas
37|2. Masukkan **email** dan **password**
38|3. Klik **"Sign In"**
39|4. Setelah login, Anda akan melihat dashboard
40|
41|#### Login dengan Google/GitHub (OAuth)
42|
43|1. Klik tombol **"Login with Google"** atau **"Login with GitHub"**
44|2. Pilih akun yang sudah terhubung
45|3. Setujui permintaan akses
46|4. Anda akan otomatis login
47|
48|---
49|
50|### 2. Dashboard & Navigasi
51|
52|Setelah login, Anda akan melihat:
53|
54|```
55|┌─────────────────────────────────────────────────────────┐
56|│  🏠 Coding Challenge    [Problems] [Leaderboard] [👤]   │
57|├────────────┬────────────────────────────────────────────┤
58|│            │                                            │
59|│  FILTER    │  🔥 Daftar Soal                           │
60|│            │                                            │
61|│  ○ All     │  ┌──────────┐ ┌──────────┐ ┌──────────┐  │
62|│  ● Easy    │  │ Two Sum  │ │ FizzBuzz │ │ Valid... │  │
63|│  ○ Medium  │  │ [Easy]   │ │ [Easy]   │ │ [Medium] │  │
64|│  ○ Hard    │  │ ✅ Solved│ │ ❌ Unsol │ │ 🔄 Attem │  │
65|│            │  └──────────┘ └──────────┘ └──────────┘  │
66|│  CATEGORY  │                                            │
67|│  □ Array   │  ┌──────────┐ ┌──────────┐ ┌──────────┐  │
68|│  □ String  │  │ Max Sub  │ │ Coin Chg │ │ Trappin  │  │
69|│  □ DP      │  │ [Easy]   │ │ [Hard]   │ │ [Hard]   │  │
70|│  □ Stack   │  └──────────┘ └──────────┘ └──────────┘  │
71|│            │                                            │
72|│  SEARCH    │  Page 1 of 2  [< Prev] [Next >]           │
73|│  [______]  │                                            │
74|└────────────┴────────────────────────────────────────────┘
75|```
76|
77|#### Navigasi Utama
78|
79|| Menu | Deskripsi |
80||------|-----------|
81|| **🏠 Home** | Dashboard dengan statistik dan soal unggulan |
82|| **📝 Problems** | Daftar semua soal dengan filter |
83|| **🏆 Leaderboard** | Peringkat pengguna |
84|| **👤 Profile** | Profil, statistik, dan riwayat submission |
85|
86|---
87|
88|### 3. Mengerjakan Soal
89|
90|#### Memilih Soal
91|
92|1. Buka halaman **"Problems"**
93|2. Gunakan **filter** untuk mencari soal:
94|   - **Difficulty**: Easy / Medium / Hard
95|   - **Category**: Array / String / DP / Stack / Graph
96|   - **Status**: Solved / Attempted / Unsolved
97|   - **Search**: Cari berdasarkan judul atau tag
98|3. Klik **card soal** untuk membuka detail
99|
100|#### Halaman Detail Soal
101|
102|```
103|┌─────────────────────────────────────────────────────────┐
104|│  ← Back to Problems                                     │
105|├────────────────────────┬────────────────────────────────┤
106|│                        │                                │
107|│  📋 Two Sum            │  💻 Code Editor                │
108|│  [Easy] [Array]        │                                │
109|│                        │  ┌──────────────────────────┐  │
110|│  DESCRIPTION           │  │ func twoSum(nums []int,  │  │
111|│  ──────────────────    │  │     target int) []int {  │  │
112|│  Given an array of     │  │     // Your code here    │  │
113|│  integers nums and an  │  │     return nil           │  │
114|│  integer target...     │  │ }                        │  │
115|│                        │  │                          │  │
116|│  EXAMPLES              │  │                          │  │
117|│  ──────────────────    │  │                          │  │
118|│  Example 1:            │  │                          │  │
119|│  Input: nums =         │  │                          │  │
120|│    [2,7,11,15],        │  │                          │  │
121|│    target = 9          │  │                          │  │
122|│  Output: [0,1]         │  │                          │  │
123|│  Explanation: nums[0]  │  │                          │  │
124|│    + nums[1] = 9       │  │                          │  │
125|│                        │  └──────────────────────────┘  │
126|│  CONSTRAINTS           │                                │
127|│  ──────────────────    │  [▶ Submit] [↺ Reset]         │
128|│  • 2 ≤ len(nums) ≤ 10⁴│                                │
129|│  • -10⁹ ≤ nums[i]     │  📊 Test Results               │
130|│                        │  ──────────────────            │
131|│  💡 HINTS              │  ⏳ Running tests...           │
132|│  ──────────────────    │                                │
133|│  Hint 1 of 3 revealed  │  ✅ Test 1: Passed (2ms)      │
134|│  ┌──────────────────┐  │  ✅ Test 2: Passed (1ms)      │
135|│  │ 💡 Use extra     │  │  ✅ Test 3: Passed (3ms)      │
136|│  │ memory           │  │  ❌ Test 4: Failed            │
137|│  │ Can you use a    │  │     Expected: [2,4]           │
138|│  │ data structure   │  │     Actual:   [1,3]           │
139|│  │ to remember...   │  │                                │
140|│  └──────────────────┘  │  📈 3/5 tests passed (60%)    │
141|│                        │                                │
142|│  [Show Next Hint]      │  ⏱ Execution: 15ms            │
143|│                        │  💾 Memory: 2.3 MB             │
144|└────────────────────────┴────────────────────────────────┘
145|```
146|
147|#### Menulis Kode
148|
149|1. **Template code** sudah tersedia di editor
150|2. Tulis solusi Anda di dalam function yang diberikan
151|3. Gunakan **CodeMirror/Monaco Editor** dengan fitur:
152|   - Syntax highlighting untuk Go
153|   - Auto-indent (4 spaces)
154|   - Line numbers
155|   - Bracket matching
156|   - Error highlighting
157|
158|#### Submit Kode
159|
160|1. Klik tombol **"▶ Submit"** atau tekan `Ctrl+Enter`
161|2. Kode akan dikirim ke server untuk dieksekusi
162|3. Hasil akan muncul secara real-time:
163|   - **⏳ Pending**: Kode sedang dalam antrean
164|   - **🔄 Running**: Kode sedang dieksekusi
165|   - **✅ Accepted**: Semua test case passed!
166|   - **❌ Wrong Answer**: Output tidak sesuai expected
167|   - **⏱ Time Limit Exceeded**: Kode terlalu lambat
168|   - **💥 Runtime Error**: Error saat eksekusi
169|   - **🔧 Compilation Error**: Syntax error
170|
171|#### Reset Kode
172|
173|- Klik tombol **"↺ Reset"** untuk mengembalikan ke template awal
174|- Atau tekan `Ctrl+R`
175|
176|---
177|
178|### 4. Sistem Hint (3 Level)
179|
180|Setiap soal memiliki **3 level hint** yang bisa dibuka secara progresif:
181|
182|| Level | Isi | Penalti Score |
183||-------|-----|---------------|
184|| **Hint 1** | Petunjuk umum (approach/pattern) | -10% |
185|| **Hint 2** | Petunjuk teknis (subproblem breakdown) | -20% |
186|| **Hint 3** | Petunjuk lanjutan (pseudocode/solusi hampir lengkap) | -30% |
187|
188|#### Cara Menggunakan Hint
189|
190|1. Scroll ke bagian **"Hints"** di halaman soal
191|2. Klik tombol **"Show Next Hint"**
192|3. Baca konfirmasi penalti score
193|4. Klik **"Yes, show hint"** untuk konfirmasi
194|5. Hint akan muncul di panel
195|
196|#### Contoh Hint
197|
198|**Hint 1 (General Approach):**
199|> 💡 **Use extra memory**
200|> Can you use a data structure to remember what you've seen so far?
201|
202|**Hint 2 (Technical Breakdown):**
203|> 💡 **Hash map approach**
204|> For each element x, check if (target - x) exists in a map you've built.
205|
206|**Hint 3 (Advanced/Pseudocode):**
207|> 💡 **One-pass hash map**
208|> ```
209|> Iterate through nums once. For each num:
210|> 1. Compute complement = target - num
211|> 2. If complement in current map, return [map[complement], i]
212|> 3. Otherwise add num → i to map
213|> ```
214|
215|---
216|
217|### 5. Leaderboard
218|
219|#### Melihat Peringat
220|
221|1. Buka halaman **"Leaderboard"**
222|2. Pilih tab:
223|   - **Weekly**: Peringkat minggu ini
224|   - **Monthly**: Peringkat bulan ini
225|   - **All Time**: Peringkat keseluruhan
226|3. Gunakan filter:
227|   - **Global**: Semua pengguna
228|   - **Friends**: Teman saja
229|
230|#### Sistem Ranking
231|
232|| Rank | Badge | Kriteria |
233||------|-------|----------|
234|| 1 | 🥇 Gold | Top 1% |
235|| 2 | 🥈 Silver | Top 5% |
236|| 3 | 🥉 Bronze | Top 10% |
237|| 4-10 | ⭐ Star | Top 25% |
238|| 11+ | 📝 Normal | Sisanya |
239|
240|#### Perhitungan Score
241|
242|```
243|Score = (Test Cases Passed / Total Test Cases) × Max Score × Hint Penalty
244|
245|Contoh:
246|- Test Cases Passed: 4/5
247|- Max Score: 100
248|- Hint Penalty: 0.9 (Hint 1 digunakan)
249|- Score = (4/5) × 100 × 0.9 = 72
250|```
251|
252|---
253|
254|### 6. Profil & Statistik
255|
256|#### Melihat Profil
257|
258|1. Klik **avatar/nama** di pojok kanan atas
259|2. Pilih **"Profile"**
260|
261|#### Statistik yang Tersedia
262|
263|| Statistik | Deskripsi |
264||-----------|-----------|
265|| **Total Solved** | Jumlah soal yang berhasil diselesaikan |
266|| **Total Submissions** | Jumlah total submission |
267|| **Acceptance Rate** | Persentase submission yang accepted |
268|| **Current Rating** | Rating ELO saat ini |
269|| **Streak** | Jumlah hari berturut-turut mengerjakan soal |
270|| **Problems by Category** | Breakdown soal per kategori |
271|| **Problems by Difficulty** | Breakdown soal per difficulty |
272|
273|#### Riwayat Submission
274|
275|1. Buka halaman **"Profile"**
276|2. Klik tab **"Submissions"**
277|3. Lihat daftar submission dengan filter:
278|   - **Status**: Accepted / Wrong Answer / Error
279|   - **Problem**: Filter per soal
280|   - **Date Range**: Filter per tanggal
281|
282|---
283|
284|## 👨‍💼 Untuk Admin
285|
286|### 1. Login Admin
287|
288|1. Login dengan akun admin (role: `admin`)
289|2. Anda akan melihat menu tambahan **"Admin Panel"** di sidebar
290|
291|### 2. Manajemen Soal
292|
293|#### Membuat Soal Baru
294|
295|1. Buka **Admin Panel** → **Problems** → **Create New**
296|2. Isi form:
297|   - **Title**: Judul soal
298|   - **Slug**: URL-friendly identifier (auto-generate)
299|   - **Description**: Deskripsi soal (Markdown)
300|   - **Difficulty**: Easy / Medium / Hard
301|   - **Category**: Pilih kategori
302|   - **Tags**: Tambah tag (comma-separated)
303|   - **Time Limit**: Dalam detik (1-5)
304|   - **Memory Limit**: Dalam MB (256-1024)
305|   - **Max Score**: Poin maksimal (default: 100)
306|   - **Function Name**: Nama function (untuk function-based)
307|   - **Return Type**: Tipe return (untuk function-based)
308|   - **Template Code**: Starter code untuk user
309|   - **Constraints**: Batasan soal (array)
310|
311|3. Tambahkan **Test Cases**:
312|   - **Input**: Input test case
313|   - **Expected Output**: Output yang diharapkan
314|   - **Is Hidden**: Centang jika hidden test case
315|   - **Is Sample**: Centang jika contoh (ditampilkan di deskripsi)
316|   - **Weight**: Bobot scoring
317|
318|4. Tambahkan **Hints**:
319|   - **Level 1**: Petunjuk umum
320|   - **Level 2**: Petunjuk teknis
321|   - **Level 3**: Petunjuk lanjutan
322|
323|5. Tambahkan **Solution** (hidden dari user):
324|   - **Code**: Kode solusi
325|   - **Approach**: Penjelasan pendekatan
326|   - **Time Complexity**: Kompleksitas waktu
327|   - **Space Complexity**: Kompleksitas ruang
328|
329|6. Klik **"Publish"** untuk mempublikasikan atau **"Save Draft"**
330|
331|#### Edit Soal
332|
333|1. Buka **Admin Panel** → **Problems**
334|2. Cari soal yang ingin diedit
335|3. Klik ikon **✏️ Edit**
336|4. Ubah field yang diperlukan
337|5. Klik **"Update"**
338|
339|#### Hapus Soal
340|
341|1. Buka **Admin Panel** → **Problems**
342|2. Cari soal yang ingin dihapus
343|3. Klik ikon **🗑️ Delete**
344|4. Konfirmasi penghapusan
345|
346|### 3. Manajemen User
347|
348|#### Melihat Daftar User
349|
350|1. Buka **Admin Panel** → **Users**
351|2. Lihat tabel user dengan kolom:
352|   - Username, Email, Role, Rating, Solved, Joined
353|
354|#### Edit User
355|
356|1. Klik ikon **✏️ Edit** pada user
357|2. Ubah field:
358|   - **Role**: user / admin / moderator
359|   - **Rating**: Rating ELO
360|   - **Status**: active / banned
361|
362|#### Ban/Unban User
363|
364|1. Klik ikon **🚫 Ban** pada user
365|2. Masukkan alasan ban
366|3. Klik **"Confirm"**
367|
368|### 4. Monitoring
369|
370|#### Dashboard Admin
371|
372|1. Buka **Admin Panel** → **Dashboard**
373|2. Lihat statistik:
374|   - **Total Users**: Jumlah pengguna terdaftar
375|   - **Active Users (24h)**: Pengguna aktif 24 jam terakhir
376|   - **Total Submissions**: Jumlah total submission
377|   - **Submissions Today**: Submission hari ini
378|   - **Server Health**: Status semua service
379|
380|#### Lihat Logs
381|
382|1. Buka **Admin Panel** → **Logs**
383|2. Filter logs:
384|   - **Level**: DEBUG / INFO / WARN / ERROR
385|   - **Service**: Filter per service
386|   - **Time Range**: Filter per waktu
387|
388|---
389|
390|## 🔌 API Reference
391|
392|### Base URL
393|
394|```
395|Development: http://localhost:9100/api
396|Production:  https://api.codingchallenge.com/api
397|```
398|
399|### Authentication
400|
401|Semua endpoint (kecuali auth) memerlukan JWT token di header:
402|
403|```
404|Authorization: Bearer ***
405|```
406|
407|### Endpoints
408|
409|#### Auth Service (Port 8081)
410|
411|| Method | Endpoint | Deskripsi | Auth |
412||--------|----------|-----------|------|
413|| POST | `/auth/register` | Registrasi akun baru | ❌ |
414|| POST | `/auth/login` | Login | ❌ |
415|| POST | `/auth/refresh` | Refresh access token | ❌ |
416|| POST | `/auth/logout` | Logout | ✅ |
417|| GET | `/auth/me` | Get current user profile | ✅ |
418|| PUT | `/auth/me` | Update profile | ✅ |
419|| POST | `/auth/change-password` | Ubah password | ✅ |
420|
421|#### Problem Service (Port 8082)
422|
423|| Method | Endpoint | Deskripsi | Auth |
424||--------|----------|-----------|------|
425|| GET | `/api/problems` | List semua soal | ❌ |
426|| GET | `/api/problems/:id` | Detail soal | ❌ |
427|| GET | `/api/problems/:id/template` | Get template code | ❌ |
428|| POST | `/api/problems` | Buat soal baru | ✅ Admin |
429|| PUT | `/api/problems/:id` | Update soal | ✅ Admin |
430|| DELETE | `/api/problems/:id` | Hapus soal | ✅ Admin |
431|
432|#### Execution Service (Port 8083)
433|
434|| Method | Endpoint | Deskripsi | Auth |
435||--------|----------|-----------|------|
436|| POST | `/api/submissions` | Submit kode | ✅ |
437|| GET | `/api/submissions/:id` | Get submission status | ✅ |
438|| GET | `/api/submissions/user/:userId` | Riwayat submission | ✅ |
439|| POST | `/api/validate` | Validasi syntax (tanpa eksekusi) | ✅ |
440|
441|#### Leaderboard Service (Port 8084)
442|
443|| Method | Endpoint | Deskripsi | Auth |
444||--------|----------|-----------|------|
445|| GET | `/api/leaderboard/weekly` | Peringkat minggu ini | ❌ |
446|| GET | `/api/leaderboard/monthly` | Peringkat bulan ini | ❌ |
447|| GET | `/api/leaderboard/all-time` | Peringkat keseluruhan | ❌ |
448|| GET | `/api/users/:id/stats` | Statistik user | ❌ |
449|
450|#### Hint Service (Port 8085)
451|
452|| Method | Endpoint | Deskripsi | Auth |
453||--------|----------|-----------|------|
454|| GET | `/api/problems/:id/hints` | Get hints untuk soal | ✅ |
455|| POST | `/api/problems/:id/hints/:hintId/use` | Gunakan hint | ✅ |
456|
457|---
458|
459|## 📝 Contoh Kode & Submission
460|
461|### Contoh 1: Two Sum (Easy)
462|
463|**Soal:** Diberikan array `nums` dan `target`, kembalikan indeks dua angka yang jumlahnya sama dengan `target`.
464|
465|**Template:**
466|```go
467|func twoSum(nums []int, target int) []int {
468|    // Your code here
469|    return nil
470|}
471|```
472|
473|**Solusi:**
474|```go
475|func twoSum(nums []int, target int) []int {
476|    seen := make(map[int]int)
477|    for i, num := range nums {
478|        if j, ok := seen[target-num]; ok {
479|            return []int{j, i}
480|        }
481|        seen[num] = i
482|    }
483|    return nil
484|}
485|```
486|
487|**Test Cases:**
488|| Input | Expected Output |
489||-------|-----------------|
490|| `[2,7,11,15], 9` | `[0,1]` |
491|| `[3,2,4], 6` | `[1,2]` |
492|| `[3,3], 6` | `[0,1]` |
493|
494|### Contoh 2: FizzBuzz (Easy)
495|
496|**Template:**
497|```go
498|func fizzBuzz(n int) []string {
499|    // Your code here
500|    return nil
501|}
502|```
503|
504|**Solusi:**
505|```go
506|func fizzBuzz(n int) []string {
507|    result := make([]string, n)
508|    for i := 1; i <= n; i++ {
509|        switch {
510|        case i%15 == 0:
511|            result[i-1] = "FizzBuzz"
512|        case i%3 == 0:
513|            result[i-1] = "Fizz"
514|        case i%5 == 0:
515|            result[i-1] = "Buzz"
516|        default:
517|            result[i-1] = strconv.Itoa(i)
518|        }
519|    }
520|    return result
521|}
522|```
523|
524|### Contoh 3: Valid Parentheses (Medium)
525|
526|**Template:**
527|```go
528|func isValid(s string) bool {
529|    // Your code here
530|    return false
531|}
532|```
533|
534|**Solusi:**
535|```go
536|func isValid(s string) bool {
537|    stack := []rune{}
538|    pairs := map[rune{rune{')': '(', ']': '[', '}': '{'}
539|    
540|    for _, ch := range s {
541|        switch ch {
542|        case '(', '[', '{':
543|            stack = append(stack, ch)
544|        case ')', ']', '}':
545|            if len(stack) == 0 || stack[len(stack)-1] != pairs[ch] {
546|                return false
547|            }
548|            stack = stack[:len(stack)-1]
549|        }
550|    }
551|    return len(stack) == 0
552|}
553|```
554|
555|### Contoh API Request
556|
557|#### Register
558|```bash
559|curl -X POST http://localhost:9100/auth/register \
560|  -H "Content-Type: application/json" \
561|  -d '{
562|    "username": "johndoe",
563|    "email": "john@example.com",
564|    "password": "SecurePass123",
565|    "full_name": "John Doe"
566|  }'
567|```
568|
569|#### Login
570|```bash
571|curl -X POST http://localhost:9100/auth/login \
572|  -H "Content-Type: application/json" \
573|  -d '{
574|    "email": "john@example.com",
575|    "password": "SecurePass123"
576|  }'
577|```
578|
579|Response:
580|```json
581|{
582|  "data": {
583|    "access_token": "eyJhbG...NiIs...",
584|    "refresh_token": "eyJhbG...NiIs...",
585|    "user": {
586|      "id": "550e8400-e29b-41d4-a716-446655440000",
587|      "username": "johndoe",
588|      "email": "john@example.com",
589|      "rating": 1200
590|    }
591|  }
592|}
593|```
594|
595|#### Submit Code
596|```bash
597|curl -X POST http://localhost:9100/api/problems/two-sum/run \
598|  -H "Content-Type: application/json" \
599|  -H "Authorization: Bearer eyJhbG......" \
600|  -d '{
601|    "code": "func twoSum(nums []int, target int) []int {\n    seen := make(map[int]int)\n    for i, num := range nums {\n        if j, ok := seen[target-num]; ok {\n            return []int{j, i}\n        }\n        seen[num] = i\n    }\n    return nil\n}"
602|  }'
603|```
604|
605|Response:
606|```json
607|{
608|  "data": {
609|    "success": true,
610|    "test_results": [
611|      {"name": "test_1", "passed": true, "expected": "[0,1]", "actual": "[0,1]"},
612|      {"name": "test_2", "passed": true, "expected": "[1,2]", "actual": "[1,2]"},
613|      {"name": "test_3", "passed": true, "expected": "[0,1]", "actual": "[0,1]"},
614|      {"name": "test_4", "passed": true, "expected": "[2,4]", "actual": "[2,4]"},
615|      {"name": "test_5", "passed": true, "expected": "[0,3]", "actual": "[0,3]"}
616|    ],
617|    "passed_count": 5,
618|    "total_count": 5,
619|    "execution_time_ms": 15,
620|    "memory_used_kb": 2300
621|  }
622|}
623|```
624|
625|---
626|
627|## ⌨️ Keyboard Shortcuts
628|
629|| Shortcut | Aksi |
630||----------|------|
631|| `Ctrl + Enter` | Submit kode |
632|| `Ctrl + R` | Reset kode ke template |
633|| `Ctrl + S` | Simpan draft (localStorage) |
634|| `Ctrl + /` | Toggle keyboard shortcuts help |
635|| `Ctrl + D` | Toggle dark/light theme |
636|| `Esc` | Close modal/panel |
637|| `Tab` | Indent (4 spaces) |
638|| `Shift + Tab` | Outdent |
639|
640|---
641|
642|## ❓ FAQ
643|
644|### Umum
645|
646|**Q: Bahasa pemrograman apa yang didukung?**
647|A: Saat ini hanya **Golang** yang didukung. Bahasa lain akan ditambahkan di masa depan.
648|
649|**Q: Berapa lama waktu eksekusi kode?**
650|A: Tergantung difficulty:
651|- **Easy**: 1 detik
652|- **Medium**: 2 detik
653|- **Hard**: 5 detik
654|
655|**Q: Berapa batas memori?**
656|A: Tergantung difficulty:
657|- **Easy**: 256 MB
658|- **Medium**: 512 MB
659|- **Hard**: 1 GB
660|
661|**Q: Apa itu "hidden test cases"?**
662|A: Test case yang tidak ditampilkan di deskripsi soal. Kode Anda harus lulus semua test case (termasuk hidden) untuk mendapatkan status "Accepted".
663|
664|**Q: Bagaimana sistem scoring?**
665|A: Score dihitung dari:
666|```
667|Score = (Test Cases Passed / Total Test Cases) × Max Score × Hint Penalty
668|```
669|
670|**Q: Apa penalti menggunakan hint?**
671|A:
672|- Hint 1: -10% score
673|- Hint 2: -20% score
674|- Hint 3: -30% score
675|
676|### Troubleshooting
677|
678|**Q: Submission saya "Pending" terus?**
679|A: Kemungkinan worker sedang sibuk. Tunggu beberapa detik atau coba lagi.
680|
681|**Q: "Time Limit Exceeded" - apa yang harus dilakukan?**
682|A: Optimasi algoritma Anda. Cari pendekatan dengan kompleksitas waktu yang lebih baik.
683|
684|**Q: "Runtime Error" - bagaimana debug?**
685|A: Periksa:
686|- Array index out of bounds
687|- Division by zero
688|- Nil pointer dereference
689|- Infinite recursion
690|
691|**Q: "Compilation Error" - kenapa?**
692|A: Periksa:
693|- Syntax error (missing bracket, semicolon, dll)
694|- Import yang tidak digunakan
695|- Tipe data yang tidak cocok
696|
697|**Q: Kode saya hilang saat refresh?**
698|A: Kode tersimpan di localStorage per problem. Jika hilang, coba cek apakah localStorage browser Anda aktif.
699|
700|### Kontak & Support
701|
702|- **Email**: support@codingchallenge.com
703|- **Discord**: https://discord.gg/codingchallenge
704|- **GitHub Issues**: https://github.com/codingchallenge/platform/issues
705|
706|---
707|
708|## 🎯 Tips & Tricks
709|
710|### Untuk Pemula
711|
712|1. **Mulai dari Easy** — Kerjakan soal Easy terlebih dahulu untuk memahami format
713|2. **Baca constraints** — Constraints memberi hint tentang kompleksitas yang dibutuhkan
714|3. **Gunakan hints bijak** — Coba selesaikan sendiri sebelum menggunakan hint
715|4. **Pelajari pattern** — Banyak soal menggunakan pattern yang sama (two pointers, sliding window, dll)
716|
717|### Untuk Intermediate
718|
719|1. **Optimasi waktu** — Selalu hitung time complexity solusi Anda
720|2. **Optimasi memori** — Cari cara mengurangi space complexity
721|3. **Edge cases** — Test dengan input kosong, single element, nilai negatif
722|4. **Baca solusi orang lain** — Pelajari approach yang berbeda
723|
724|### Untuk Advanced
725|
726|1. **DP problems** — Pahami optimal substructure dan overlapping subproblems
727|2. **Graph problems** — Kuasai BFS, DFS, Dijkstra, dan Union-Find
728|3. **Tree problems** — Pahami traversal (inorder, preorder, postorder)
729|4. **Contest** — Ikuti weekly contest untuk berlatih under pressure
730|
731|---
732|
733|> **Selamat mengerjakan! 🚀**
734|> Ingat: "The only way to learn a new programming language is by writing programs in it." — Dennis Ritchie
735|