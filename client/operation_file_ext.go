package client

import (
	"github.com/kernel-ai/koscore/client/entity"
	"github.com/kernel-ai/koscore/client/packets/oidb"
)

// 获取群文件系统信息
func (m *QQClient) GetGroupFileSystemInfo(groupUin uint64) (*entity.GroupFileSystemInfo, error) {
	pkt, err := oidb.BuildGroupFileCountReq(groupUin)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	count, err := oidb.ParseGroupFileCountResp(pkt.Data)
	if err != nil {
		return nil, err
	}

	pkt, err = oidb.BuildGroupFileSpaceReq(groupUin)
	if err != nil {
		return nil, err
	}
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
		return nil, err
	}
	space, err := oidb.ParseGroupFileSpaceResp(pkt.Data)
	if err != nil {
		return nil, err
	}

	return &entity.GroupFileSystemInfo{
		GroupUin:   groupUin,
		FileCount:  count.FileCount,
		LimitCount: count.LimitCount,
		TotalSpace: space.TotalSpace,
		UsedSpace:  space.UsedSpace,
	}, nil
}

// 获取群目录指定文件夹列表
func (m *QQClient) ListGroupFilesByFolder(groupUin uint64, target_dir string) ([]*entity.GroupFile, []*entity.GroupFolder, error) {
	var startIndex uint32
	var fileCount uint32 = 20
	var files []*entity.GroupFile
	var folders []*entity.GroupFolder
	for {
		pkt, err := oidb.BuildGroupFileListReq(groupUin, target_dir, startIndex, fileCount)
		if err != nil {
			return files, folders, err
		}
		if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil {
			return files, folders, err
		}
		res, err := oidb.ParseGroupFileListResp(pkt.Data)
		if err != nil {
			return files, folders, err
		}
		for _, fe := range res.List.Items {
			if fe.FileInfo != nil {
				files = append(files, &entity.GroupFile{
					GroupUin:      groupUin,
					FileId:        fe.FileInfo.FileId,
					FileName:      fe.FileInfo.FileName,
					BusId:         fe.FileInfo.BusId,
					FileSize:      fe.FileInfo.FileSize,
					UploadTime:    fe.FileInfo.UploadedTime,
					DeadTime:      fe.FileInfo.ExpireTime,
					ModifyTime:    fe.FileInfo.ModifiedTime,
					DownloadTimes: fe.FileInfo.DownloadedTimes,
					Uploader:      fe.FileInfo.UploaderUin,
					UploaderName:  fe.FileInfo.UploaderName,
				})
			}
			if fe.FolderInfo != nil {
				folders = append(folders, &entity.GroupFolder{
					GroupUin:       groupUin,
					FolderId:       fe.FolderInfo.FolderId,
					FolderName:     fe.FolderInfo.FolderName,
					CreateTime:     fe.FolderInfo.CreateTime,
					Creator:        fe.FolderInfo.CreatorUin,
					CreatorName:    fe.FolderInfo.CreatorName,
					TotalFileCount: fe.FolderInfo.TotalFileCount,
				})
			}
		}
		startIndex += fileCount
		if res.List.IsEnd {
			break
		}
	}
	return files, folders, nil
}

// 获取群根目录文件列表
func (m *QQClient) ListGroupRootFiles(groupUin uint64) ([]*entity.GroupFile, []*entity.GroupFolder, error) {
	return m.ListGroupFilesByFolder(groupUin, "/")
}

/*

// 重命名群文件
func (c *QQClient) RenameGroupFile(groupUin uint64, fileId string, parent_dir string, newFileName string) error {
	pkt, err := oidb.BuildGroupFileRenameReq(groupUin, fileId, parent_dir, newFileName)
	if   err != nil { return err }
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil { return err }
	return oidb.ParseGroupFileRenameResp(pkt.Data)
}

// 创建群文件夹
func (c *QQClient) CreateGroupFolder(groupUin uint64, target_dir string, dir_name string) error {
	pkt, err := oidb.BuildGroupFolderCreateReq(groupUin, target_dir, dir_name)
	if   err != nil { return err }
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil { return err }
	return oidb.ParseGroupFolderCreateResp(pkt.Data)
}

// 重命名群文件夹
func (c *QQClient) RenameGroupFolder(groupUin uint64, dir_id string, new_dir_name string) error {
	pkt, err := oidb.BuildGroupFolderRenameReq(groupUin, dir_id, new_dir_name)
	if   err != nil { return err }
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil { return err }
	return oidb.ParseGroupFolderRenameResp(pkt.Data)
}

// 删除群文件夹
func (c *QQClient) DeleteGroupFolder(groupUin uint64, dir_id string) error {
	pkt, err := oidb.BuildGroupFolderDeleteReq(groupUin, dir_id)
	if   err != nil { return err }
	if pkt, err = m.sendOidbPacketAndWait(pkt); err != nil { return err }
	return oidb.ParseGroupFolderDeleteResp(pkt.Data)
}
*/
