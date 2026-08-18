package onebotfilter

import (
	"bytes"
	"encoding/json"
)

// NormalizeFileSegments 兼容某些不携带 file_id / file 字段的实现：
// 当 file 消息段只带 id 时，自动补上 file_id 与 file，以便下游按
// OneBot 11 常见方式取用文件标识。
func NormalizeFileSegments(raw []byte) []byte {
	// 快速路径：只处理带 file 段的消息事件，其余消息原样放行
	if !bytes.Contains(raw, []byte(`"type":"file"`)) {
		return raw
	}
	var event struct {
		Message json.RawMessage `json:"message"`
	}
	if err := json.Unmarshal(raw, &event); err != nil {
		return raw
	}
	// 只处理数组格式的 message（CQ string 或通知事件不改写）
	if len(event.Message) == 0 || event.Message[0] != '[' {
		return raw
	}
	var segments []map[string]json.RawMessage
	if err := json.Unmarshal(event.Message, &segments); err != nil {
		return raw
	}
	changed := false
	for i, seg := range segments {
		var segType string
		if err := json.Unmarshal(seg["type"], &segType); err != nil || segType != "file" {
			continue
		}
		var data map[string]json.RawMessage
		if err := json.Unmarshal(seg["data"], &data); err != nil {
			continue
		}
		var id string
		if err := json.Unmarshal(data["id"], &id); err != nil || id == "" {
			continue
		}
		segChanged := false
		if _, ok := data["file_id"]; !ok {
			data["file_id"], _ = json.Marshal(id)
			segChanged = true
		}
		if _, ok := data["file"]; !ok {
			data["file"], _ = json.Marshal(id)
			segChanged = true
		}
		if segChanged {
			segData, err := json.Marshal(data)
			if err != nil {
				return raw
			}
			segments[i]["data"] = segData
			changed = true
		}
	}
	if !changed {
		return raw
	}
	msgBytes, err := json.Marshal(segments)
	if err != nil {
		return raw
	}
	var full map[string]json.RawMessage
	if err := json.Unmarshal(raw, &full); err != nil {
		return raw
	}
	full["message"] = msgBytes
	out, err := json.Marshal(full)
	if err != nil {
		return raw
	}
	return out
}
