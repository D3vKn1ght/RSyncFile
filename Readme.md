# RSyncFile

RSyncFile là một ứng dụng client-server đơn giản sử dụng Go và FastAPI để đồng bộ hóa tệp giữa client và server. Ứng dụng bao gồm chức năng tải lên và tải xuống tệp giữa client và server, đồng thời tự động kiểm tra và phát hiện thay đổi trong thư mục được giám sát.

## Cấu trúc dự án

- **Client**: Được viết bằng Go, nhiệm vụ chính là theo dõi các thay đổi trong thư mục và tải lên server, đồng thời tải xuống các tệp từ server về một thư mục định trước.
- **Server**: Được xây dựng bằng FastAPI, chịu trách nhiệm xử lý các yêu cầu tải lên, tải xuống và xóa tệp từ client.

## Cách sử dụng

### 1. Cấu hình Client

Tạo một tệp `config.json` với nội dung sau:

```jsx
{
    "url": "http://192.168.221.209:7888",
    "folder_to_watch": "C:\\RSyncFolder",
    "folder_to_receive": "C:\\Users\\Administrator\\Desktop\\test",
    "time_to_wait": 10
}
```

- `url`: URL của server FastAPI.
- `folder_to_watch`: Thư mục mà client sẽ theo dõi để phát hiện thay đổi.
- `folder_to_receive`: Thư mục nơi client sẽ lưu các tệp tải xuống từ server.
- `time_to_wait`: Thời gian đồng bộ giữa client và server (giây).

### 2. Chạy Server

1. Build image:

```jsx
docker build -t rsyncfile .
```

1. Chạy server FastAPI:

```jsx
docker run -p 7888:7888 -v /path_folder_dupload:/upload -v /path_folder_download:/download rsyncfile
```

### 3. Chạy Client

1. Biên dịch mã Go (nếu chưa có tệp thực thi):\

```jsx
go build -o ClientRsyncFile.exe

```

1. Chạy Client:

```jsx
./ClientRsyncFile.exe
```

Client sẽ bắt đầu theo dõi thư mục và thực hiện đồng bộ hóa với server.

## Các chức năng chính

### Tải lên tệp

- Client tự động phát hiện thay đổi trong `folder_to_watch` và tải lên server.

### Tải xuống tệp

- Client sẽ tải xuống tất cả các tệp có trong thư mục `/download` của server và lưu vào `folder_to_receive`.

### Xóa tệp

- Client có thể xóa tệp từ server nếu tệp không còn tồn tại trong `folder_to_watch`.

## Góp ý và báo lỗi

Nếu bạn gặp bất kỳ vấn đề nào hoặc có đề xuất cải tiến, vui lòng mở một Issue hoặc Pull Request trên GitHub.