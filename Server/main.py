from fastapi import FastAPI, File, UploadFile, HTTPException
import os

app = FastAPI()

# Đảm bảo thư mục /upload tồn tại
UPLOAD_FOLDER = "/upload"
os.makedirs(UPLOAD_FOLDER, exist_ok=True)

@app.post("/uploadfile/")
async def upload_file(file: UploadFile = File(...)):
    try:
        if not file:
            raise HTTPException(status_code=400, detail="No file uploaded")

        tmp_file_location = os.path.join(UPLOAD_FOLDER, f"{file.filename}.tmp")

        with open(tmp_file_location, "wb") as tmp_file:
            content = await file.read()
            tmp_file.write(content)

        if not os.path.exists(tmp_file_location) or os.path.getsize(tmp_file_location) == 0:
            raise HTTPException(status_code=500, detail="Failed to save the uploaded file")

        final_file_location = os.path.join(UPLOAD_FOLDER, file.filename)
        os.rename(tmp_file_location, final_file_location)

        return {"filename": file.filename, "status": "File uploaded and renamed successfully"}

    except HTTPException as http_exc:
        raise http_exc
    except Exception as e:
        if os.path.exists(tmp_file_location):
            os.remove(tmp_file_location)
        raise HTTPException(status_code=500, detail=f"An error occurred: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=7888)
