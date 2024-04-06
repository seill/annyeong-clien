### 대하여
`Golang`에서 `Selenium`을 이용하여 `clien.net` 내가 작성한 글과 그 글에 달려있는 댓글을 `json` 포멧으로 저장합니다. 글에 포함된 이미지파일은 `images` 디렉토리에 저장됩니다.
본 프로젝트는 `Silicon Mac`에서 작성되었고, 테스트 되었습니다.
---
### 사용하기전
- `Golang`이 설치되어있어야합니다. [다운로드](https://golang.org/dl/)
- `Chrome Driver`가 설치되어있어야합니다. [다운로드](https://chromedriver.chromium.org/downloads)
- 설치된 `Chrome Driver`가 본 프로젝트 디렉토리에 복사되어 있어야합니다.
- `Chrome Driver` 관련 실행 오류시 다음 명령어로 해결이 가능합니다.
```shell
xattr -d com.apple.quarantine chromedriver
```
---
### 사용법
- `Makefile.template` 파일을 `Makefile`로 이름을 변경하거나 복사합니다.
```shell
cp Makefile.template Makefile
```
- 편집기로 `Makefile`을 열어서 `<id>` `<password>` 부분을 수정합니다.
- `id`, `password`는 실행되는 웹브라우저에서 입력할 수도 있으니 삭제해도 됩니다.
- Build 하기
```shell
make build
```
- 내가 쓴 글 저장하기 (`내가 쓴 댓글 저장은 아직 구현 안됨`)
```shell
make archive
```
- 내가 쓴 글 삭제하기 (`아직 구현 안됨`)
```
make delete
```
---
