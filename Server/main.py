from fastapi import FastAPI, File, UploadFile, HTTPException, Header
import os
import gzip
import shutil

app = FastAPI()

UPLOAD_FOLDER = "/upload"
os.makedirs(UPLOAD_FOLDER, exist_ok=True)

@app.get("/")
async def root():
    return {"message": "Hello World"}

@app.post("/uploadfile/")
async def upload_file(file: UploadFile = File(...), filename: str = Header(...)):
    try:
        if not file:
            raise HTTPException(status_code=400, detail="No file uploaded")

        tmp_file_location = os.path.join(UPLOAD_FOLDER, f"{filename}.tmp")
        os.makedirs(os.path.dirname(tmp_file_location), exist_ok=True)
        with open(tmp_file_location, "wb") as tmp_file:
            content = await file.read()
            tmp_file.write(content)

        gz_file_location = tmp_file_location
        final_file_location = gz_file_location.rstrip(".tmp").rstrip(".gz")

        with gzip.open(gz_file_location, 'rb') as f_in:
            with open(final_file_location, 'wb') as f_out:
                shutil.copyfileobj(f_in, f_out)

        os.remove(gz_file_location)

        return {"filename": final_file_location, "status": "File uploaded, extracted, and renamed successfully"}

    except HTTPException as http_exc:
        raise http_exc
    except Exception as e:
        if os.path.exists(tmp_file_location):
            os.remove(tmp_file_location)
        raise HTTPException(status_code=500, detail=f"An error occurred: {str(e)}")

if __name__ == "__main__":
    import uvicorn
    uvicorn.run(app, host="0.0.0.0", port=8000)
