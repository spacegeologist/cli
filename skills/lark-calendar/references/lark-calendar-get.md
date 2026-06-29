
# calendar +get

通过 `calendar_id` + `event_id` 获取**单个日程**的详情。只读，不修改任何数据。

## 命令

```bash
# 主日历（默认primary）
lark-cli calendar +get --event-id <event_id>

# 指定日历
lark-cli calendar +get --calendar-id <calendar_id> --event-id <event_id>

# 人类可读格式
lark-cli calendar +get --event-id <event_id> --format pretty

# 预览 API 调用，不真实执行
lark-cli calendar +get --event-id <event_id> --dry-run
```

## 参数

| 参数 | 必填 | 说明 |
|------|------|------|
| `--calendar-id <id>` | 否 | 日历 ID。省略时使用主日历 `primary` |
| `--event-id <id>` | 是 | 日程 ID。重复性日程的某次实例必须传该实例的 `event_id`，禁止使用原循环日程的 `event_id` |
| `--format` | 否 | 输出格式：`json`（默认） \| `pretty` |
| `--dry-run` | 否 | 只打印将要发起的 API 请求，不真正调用 |

## 输出字段

| 字段 | 类型 | 说明 |
|------|------|------|
| `event_id` | string | 日程 ID |
| `organizer_calendar_id` | string | 组织者日历 ID |
| `summary` | string | 日程标题 |
| `description` | string | 日程描述 |
| `start_time.datetime` | string | 开始时间，**RFC3339（设备本地时区）**；非全天日程返回 |
| `start_time.date` | string | 开始日期，仅全天日程返回 |
| `start_time.timezone` | string | 时区（IANA），仅非全天日程返回 |
| `end_time.datetime` | string | 结束时间，RFC3339；非全天日程返回 |
| `end_time.date` | string | 结束日期，仅全天日程返回。**已按"包含"语义自动回卷为最后一天** |
| `end_time.timezone` | string | 时区（IANA），仅非全天日程返回 |
| `vchat` | object | 视频会议配置（`vc_type` / `meeting_url` / `icon_type` / `description`） |
| `visibility` | string | 可见性：`default` / `public` / `private` |
| `attendee_ability` | string | 参会人权限：`none` / `can_see_others` / `can_invite_others` / `can_modify_event` |
| `free_busy_status` | string | 忙闲状态：`busy` / `free` |
| `self_rsvp_status` | string | 当前用户的 RSVP：`needs_action` / `accept` / `decline` / `tentative` / `removed` |
| `location` | object | 地点（`name` / `address` / `latitude` / `longitude`） |
| `color` | int | 日程颜色（-1 为继承） |
| `reminders[]` | array | 提醒，单元素 `{minutes}` |
| `recurrence` | string | RRULE 字符串（仅循环日程） |
| `is_exception` | bool | 是否是循环日程的例外实例 |
| `recurring_event_id` | string | 原循环日程 ID（仅例外实例） |
| `create_time` | string | 创建时间，**RFC3339（设备本地时区）** |
| `event_organizer` | object | 创建人（`user_id` / `display_name`） |
| `app_link` | string | 飞书内部唤起链接 |
| `attachments[]` | array | 附件（`file_token` / `file_size` / `name`） |
| `event_check_in` | object | 签到配置 |
| `status` | string | **仅当 `status == "cancelled"` 才返回**，非取消日程会从输出中删除 |
