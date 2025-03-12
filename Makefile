.PHONY: commit release

commit:
	@read -p "请输入提交信息: " message; \
	git add .; \
	git commit -m "$$message"

release:
	@read -p "请输入版本号(例如 v1.0.0): " version; \
	git tag -a $$version -m "Release $$version"; \
	git push origin $$version; \
	git push origin main